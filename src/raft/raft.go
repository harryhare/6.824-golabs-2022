package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, Term, isleader)
//   start agreement on a new log entry
// rf.GetState() (Term, isLeader)
//   ask a Raft for its current Term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"6.824/labgob"
	"bytes"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	//	"6.824/labgob"
	"6.824/labrpc"
)

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 2D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
//
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

type LogEntry struct {
	Term    int64
	Command interface{}
}
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()
	applyCh   chan ApplyMsg

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	logs          []LogEntry
	commitedIndex int

	term         int64
	leader       int
	vote_for     int
	vote_timeout time.Time
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	var term int64 = rf.term
	var isleader bool = (rf.leader == rf.me)
	// Your code here (2A).
	return int(term), isleader
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(rf.term)
	e.Encode(rf.vote_for)
	e.Encode(rf.logs)
	data := w.Bytes()
	rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	r := bytes.NewBuffer(data)
	d := labgob.NewDecoder(r)
	err := d.Decode(&rf.term)
	if err != nil {
		panic(err)
	}
	err = d.Decode(&rf.vote_for)
	if err != nil {
		panic(err)
	}
	err = d.Decode(&rf.logs)
	if err != nil {
		panic(err)
	}
}

//
// A service wants to switch to snapshot.  Only do so if Raft hasn't
// have more recent info since it communicate the snapshot on applyCh.
//
func (rf *Raft) CondInstallSnapshot(lastIncludedTerm int, lastIncludedIndex int, snapshot []byte) bool {

	// Your code here (2D).

	return true
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (2D).

}

type AppendEntriesArgs struct {
	Term   int64
	Sender int

	Entries     []LogEntry
	PrevIndex   int
	PrevTerm    int64
	CommitIndex int
}
type AppendEntriesReply struct {
	Ok   bool
	Term int64

	PrevIndex int //
	PrevTerm  int64
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	reply.Term = rf.term
	if rf.term > args.Term {
		reply.Ok = false
		//reply.Term = rf.term
		DPrintf("%d, %d reject AppendEntries from %d, leader %d, term too low %+v", rf.term, rf.me, args.Sender, rf.leader, args)
		return
	}
	if rf.term == args.Term && rf.leader != -1 && rf.leader != args.Sender {
		DPrintf("%d, %d reject AppendEntries from %d, leader %d, sender is not leader %+v", rf.term, rf.me, args.Sender, rf.leader, args)
		reply.Ok = false
		return
	}
	reply.Ok = true
	rf.leader = args.Sender
	rf.term = args.Term
	d := get_time_out()
	rf.vote_timeout = time.Now().Add(d)
	DPrintf("%d, %d recv AppendEntries  from %d, vote_timeout+%v, req %+v", rf.term, rf.me, args.Sender, d, args)

	//if args.PrevIndex == -1 {
	//	reply.Ok = true
	//	//over write
	//	rf.logs = rf.logs[:args.PrevIndex+1]
	//	rf.logs = append(rf.logs, args.Entries...)
	//	DPrintf("%d, %d recv AppendEntries  from %d, vote_timeout+%v, prev term %d, Entries appended", rf.term, rf.me, args.Sender, d, args.Term)
	//} else
	if len(rf.logs)-1 < args.PrevIndex {
		reply.Ok = false
		last_entry := rf.logs[len(rf.logs)-1]
		reply.PrevTerm = last_entry.Term
		reply.PrevIndex = len(rf.logs) - 1
		DPrintf("%d, %d recv AppendEntries  from %d, vote_timeout+%v, prevIndex is empty, reply %+v", rf.term, rf.me, args.Sender, d, reply)
		return
	} else {
		prev_entry := rf.logs[args.PrevIndex]
		if prev_entry.Term != args.PrevTerm {
			reply.Ok = false
			i := args.PrevIndex
			for ; i >= 0; i-- {
				if rf.logs[i].Term != prev_entry.Term {
					break
				}
			}
			reply.PrevTerm = -1
			reply.PrevIndex = i
			if i >= 0 {
				reply.PrevTerm = rf.logs[i].Term
			}
			DPrintf("%d, %d recv AppendEntries  from %d, vote_timeout+%v, prevIndex term not match local %d leader %d, reply %+v", rf.term, rf.me, args.Sender, d, prev_entry.Term, args.PrevTerm, reply)
			return
		} else {
			reply.Ok = true
			//over write
			rf.logs = rf.logs[:args.PrevIndex+1]
			rf.logs = append(rf.logs, args.Entries...)
			DPrintf("%d, %d recv AppendEntries  from %d, vote_timeout+%v, prev term %d, Entries appended", rf.term, rf.me, args.Sender, d, args.Term)

		}

	}

	if rf.commitedIndex < args.CommitIndex {
		DPrintf("%d, %d recv AppendEntries  from %d, commit index from %d to %d, len(logs) %d", rf.term, rf.me, args.Sender, rf.commitedIndex, args.CommitIndex, len(rf.logs))
		r := args.CommitIndex
		if len(rf.logs)-1 < r {
			r = len(rf.logs) - 1
		}
		rf.apply_entries(r)
	}
	rf.persist() // entry follower
}

func (rf *Raft) apply_entries(r int) {
	for i := rf.commitedIndex + 1; i <= r; i++ {
		msg := ApplyMsg{
			CommandIndex: i, // index for outside,need + 1
			Command:      rf.logs[i].Command,
			CommandValid: true,
		}

		DPrintf("%d, %d apply commit index %d:%+v,", rf.term, rf.me, i, rf.logs[i].Command)
		rf.applyCh <- msg
	}
	rf.commitedIndex = r
}

type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Sender int
	Term   int64

	LastlogTerm   int64
	LastLogIndex  int
	LastlogLength int
}

type RequestVoteReply struct {
	// Your data here (2A).
	Ok   bool
	Term int64
}

func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	reply.Ok = true
	reply.Term = rf.term
	if rf.term > args.Term {
		reply.Ok = false
		//reply.Term = rf.term
		DPrintf("%d, %d reject vote to %d, request term too low %d", rf.term, rf.me, args.Sender, args.Term)
		return
	}
	if rf.term == args.Term && rf.vote_for != args.Sender && rf.vote_for != -1 {
		reply.Ok = false
		DPrintf("%d, %d reject vote to %d, already voted for %d", rf.term, rf.me, args.Sender, rf.vote_for)
		return
	}

	lastLogIndex := len(rf.logs) - 1
	lastEntry := rf.logs[lastLogIndex]
	if lastEntry.Term > args.LastlogTerm {
		reply.Ok = false
		DPrintf("%d, %d reject vote to %d, last log term too low %d > %d", rf.term, rf.me, args.Sender, lastEntry.Term, args.LastlogTerm)
		return
	}
	if lastEntry.Term == args.LastlogTerm {
		lastLogLength := len(rf.logs)
		if lastLogLength > args.LastlogLength {
			reply.Ok = false
			DPrintf("%d, %d reject vote to %d, last log length too low %d > %d", rf.term, rf.me, args.Sender, lastLogLength, args.LastlogLength)
			return
		}
	}

	reply.Ok = true
	rf.term = args.Term
	rf.vote_for = args.Sender
	rf.leader = -1
	d := get_time_out()
	rf.vote_timeout = time.Now().Add(d)
	DPrintf("%d, %d vote to %d, vote_timeout+%v, args %+v, logs %+v", rf.term, rf.me, args.Sender, d, args, rf.logs)

	rf.persist() // for vote
}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a vote_timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}
func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// Term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	index := len(rf.logs)
	term := rf.term
	isLeader := (rf.leader == rf.me)
	if !isLeader {
		return -1, int(term), false
	}
	rf.logs = append(rf.logs, LogEntry{
		Term:    term,
		Command: command,
	})
	DPrintf("%d, %d, append entry %+v", rf.term, rf.me, rf.logs)
	//go rf.leader_send_entries() // uncomment to speed up
	return index, int(term), true
}

//
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func (rf *Raft) leader_send_entries() {
	rf.mu.Lock()
	term := rf.term

	last_commit := rf.commitedIndex
	new_commit_index := len(rf.logs) - 1

	prevIndex := last_commit
	prevTerm := rf.logs[prevIndex].Term

	arg1 := AppendEntriesArgs{
		Term:        term,
		Sender:      rf.me,
		Entries:     rf.logs[prevIndex+1 : new_commit_index+1],
		PrevIndex:   prevIndex,
		PrevTerm:    prevTerm,
		CommitIndex: last_commit,
	}
	rf.mu.Unlock()

	is_leader := true
	quota := (len(rf.peers) / 2)
	suc_ch := make(chan int, len(rf.peers))
	err_ch := make(chan int, len(rf.peers))

	DPrintf("%d, %d send entries", term, rf.me)
	for i, _ := range rf.peers {
		if i == rf.me {
			continue
		}
		go func(i int) {
			loop := true
			arg := &arg1
			for loop {
				reply := AppendEntriesReply{}
				suc := rf.sendAppendEntries(i, arg, &reply)
				if !suc {
					suc_ch <- 0 //network err
					DPrintf("%d, %d send entries, %d network error", term, rf.me, i)
					return
				}
				if reply.Term > term {
					err_ch <- 1
					rf.mu.Lock()
					if reply.Term > rf.term {
						rf.term = reply.Term
						rf.vote_for = -1
						rf.leader = -1
					}
					rf.mu.Unlock()
					return
				}
				if reply.Ok {
					loop = false
					suc_ch <- 1
					DPrintf("%d, %d send heartbeat, %d accept", term, rf.me, i)
					return
				}
				// if !reply.Ok
				rf.mu.Lock()
				pi := reply.PrevIndex
				pt := rf.logs[pi].Term
				arg = &AppendEntriesArgs{
					Term:        term,
					Sender:      rf.me,
					Entries:     rf.logs[pi+1 : new_commit_index+1],
					PrevIndex:   pi,
					PrevTerm:    pt,
					CommitIndex: rf.commitedIndex,
				}
				rf.mu.Unlock()
				DPrintf("%d, %d send heartbeat, %d reject, reply %+v, resend %+v", term, rf.me, i, reply, arg)
			}
		}(i)
	}

	counter := 0
	suc_counter := 0
	loop := true
	for loop {
		select {
		case ok := <-suc_ch:
			suc_counter += ok
			counter += 1
			if suc_counter == len(rf.peers)-1 {
				rf.mu.Lock()
				DPrintf("%d, %d leader update commited index to %d", term, rf.me, new_commit_index)
				rf.apply_entries(new_commit_index)
				rf.mu.Unlock()
				loop = false
				break
			}
			if suc_counter >= quota {
				rf.mu.Lock()
				//rf.apply_entries(new_commit_index)
				last_log := rf.logs[new_commit_index]
				if last_log.Term == term {
					DPrintf("%d, %d leader update commited index to %d", term, rf.me, new_commit_index)
					//rf.commitedIndex = new_commit_index
					rf.apply_entries(new_commit_index)
				}
				rf.mu.Unlock()
			}
			if counter == len(rf.peers)-1 {
				loop = false
				break
			}
		case <-err_ch:
			is_leader = false
			loop = false
			break
		case <-time.After(heartbeat_timeout):
			loop = false
			break
		}
	}

	DPrintf("%d, %d send heartbeat, %d reply total", term, rf.me, counter)
	if is_leader == false {
		rf.mu.Lock()
		rf.leader = -1
		rf.mu.Unlock()
	}
	rf.mu.Lock()
	rf.persist() // entry leader
	rf.mu.Unlock()
}

func (rf *Raft) leader_ticker() {
	rf.mu.Lock()
	term := rf.term
	rf.mu.Unlock()

	arg1 := AppendEntriesArgs{
		Term:        term,
		Sender:      rf.me,
		CommitIndex: rf.commitedIndex,
	}
	is_leader := true
	quota := (len(rf.peers) / 2)
	suc_ch := make(chan int, len(rf.peers))
	err_ch := make(chan int, len(rf.peers))

	DPrintf("%d, %d send heartbeat, commited index %d", term, rf.me, rf.commitedIndex)
	for i, _ := range rf.peers {
		if i == rf.me {
			continue
		}
		go func(i int) {
			reply := AppendEntriesReply{}
			suc := rf.sendAppendEntries(i, &arg1, &reply)
			if !suc {
				suc_ch <- 0 //network err
				DPrintf("%d, %d send heartbeat, %d network error", term, rf.me, i)
				return
			}
			if reply.Term > term {
				err_ch <- 1
				rf.mu.Lock()
				if reply.Term > rf.term {
					rf.term = reply.Term
					rf.vote_for = -1
					rf.leader = -1
				}
				rf.mu.Unlock()
				return
			}
			if reply.Ok {
				suc_ch <- 1
				DPrintf("%d, %d send heartbeat, %d accept", term, rf.me, i)
				return
			}
			// impossible branch
			if !reply.Ok {
				DPrintf("%d, %d send heartbeat, %d reject", term, rf.me, i)
			}

		}(i)
	}
	counter := 0
	loop := true
	for loop {
		select {
		case ok := <-suc_ch:
			counter += ok
			if counter >= quota {
				loop = false
				break
			}
		case <-err_ch:
			is_leader = false
			loop = false
			break
		case <-time.After(heartbeat_timeout):
			loop = false
			break
		}
	}

	DPrintf("%d, %d send heartbeat, %d reply total", term, rf.me, counter)
	if is_leader == false {
		rf.mu.Lock()
		rf.leader = -1
		rf.mu.Unlock()
	}
}

func (rf *Raft) vote_ticker() {
	rf.mu.Lock()
	rf.term += 1
	term := rf.term
	rf.leader = -1
	rf.vote_for = rf.me
	lastIndex := len(rf.logs) - 1
	lastLog := rf.logs[lastIndex]
	lastTerm := lastLog.Term
	lastLength := len(rf.logs)
	rf.mu.Unlock()

	arg := RequestVoteArgs{
		Sender:        rf.me,
		Term:          term,
		LastLogIndex:  lastIndex,
		LastlogTerm:   lastTerm,
		LastlogLength: lastLength,
	}
	quota := (len(rf.peers) / 2)
	suc_ch := make(chan int, len(rf.peers))
	err_ch := make(chan int, len(rf.peers))
	DPrintf("%d, %d request vote", term, rf.me)
	for i, _ := range rf.peers {
		if i == rf.me {
			continue
		}
		go func(i int) {
			reply := RequestVoteReply{}
			suc := rf.sendRequestVote(i, &arg, &reply)
			if !suc {
				suc_ch <- 0
				return
			}
			if reply.Term > term {
				err_ch <- 1
				rf.mu.Lock()
				if reply.Term > rf.term {
					rf.term = reply.Term
					rf.vote_for = -1
					rf.leader = -1
				}
				rf.mu.Unlock()
				return
			}
			if !reply.Ok {
				err_ch <- 1
				rf.mu.Lock()
				if reply.Term >= rf.term {
					rf.term = reply.Term
					rf.vote_for = -1
					rf.leader = -1
				}
				rf.mu.Unlock()
				return
			}

			if reply.Ok {
				DPrintf("%d, %d recv vote from %d %v", term, rf.me, i, true)
				suc_ch <- 1
				return
			}
		}(i)
	}

	counter := 0
	loop := true
	for loop {
		select {
		case ok := <-suc_ch:
			counter += ok
			if counter >= quota {
				loop = false
				break
			}
		case <-err_ch:
			loop = false
			break
		case <-time.After(request_vote_timeout):
			loop = false
			break
		}

	}
	DPrintf("%d, %d get vote %d", term, rf.me, counter)
	rf.mu.Lock()
	if counter < quota || term != rf.term {
		rf.mu.Unlock()
		time.Sleep(time.Duration(rand.Intn(reelect_time)) * time.Millisecond)
		return
	}
	rf.leader = rf.me
	rf.persist() // vote leader
	rf.mu.Unlock()
	DPrintf("%d, %d is leader", term, rf.me)
}

func (rf *Raft) ticker() {
	for rf.killed() == false {
		rf.mu.Lock()
		leader := rf.leader
		vote_timeout := rf.vote_timeout
		rf.mu.Unlock()

		isleader := leader == rf.me
		start := time.Now()
		if isleader {
			//rf.leader_ticker()
			rf.leader_send_entries()
			d := heartbeat_timeout - time.Now().Sub(start)
			if d > 0 {
				time.Sleep(d)
			}
			continue
		}

		d := vote_timeout.Sub(start)
		if d > 0 {
			time.Sleep(d)
			continue
		}

		rf.vote_ticker()
	}
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//

const heartbeat_timeout = 100 * time.Millisecond    // should be small enough to avoid followers timeout
const request_vote_timeout = 200 * time.Millisecond //  should be big enough to allow followers' vote return 100* sqrt(len(rf.peer))
const min_timeout = 250                             // 3* heart beat
const max_timeout = 500                             // max-min > RTT
const reelect_time = 200                            // should be random

func get_time_out() time.Duration {
	return time.Duration(rand.Intn(max_timeout-min_timeout)+min_timeout) * time.Millisecond
}

func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	rand.Seed(int64(me))

	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.applyCh = applyCh
	rf.dead = 0
	rf.leader = -1
	rf.vote_for = -1

	rf.commitedIndex = 0 //last commited index

	// Your initialization code here (2A, 2B, 2C).

	rf.vote_timeout = time.Now().Add(get_time_out())
	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())
	if rf.logs == nil || len(rf.logs) == 0 {
		rf.logs = []LogEntry{
			{
				Term:    -1,
				Command: "Sentinel",
			},
		}
	}

	DPrintf("== %d, start term %d, vote_for %d, logs %v", rf.me, rf.term, rf.vote_for, rf.logs)
	// start ticker goroutine to start elections
	go rf.ticker()

	return rf
}
