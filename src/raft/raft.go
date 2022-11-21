package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

//11.20 Pass lab2A all test cases!

import (
	//	"bytes"
	"sync"
	"sync/atomic"
	"time"

	//	"6.824/labgob"
	"6.824/labrpc"
)

// Status 节点的角色
type Status int

// VoteState 投票的状态 2A
type VoteState int

// AppendEntriesState 追加日志的状态 2A 2B
type AppendEntriesState int

// HeartBeatTimeout 定义一个全局心跳超时时间
var HeartBeatTimeout = 120 * time.Millisecond

// 枚举节点的类型：跟随者、竞选者、领导者
const (
	Follower Status = iota
	Candidate
	Leader
)

const (
	Normal VoteState = iota //投票过程正常
	Killed                  //Raft节点已终止
	Expire                  //投票(消息\竞选者）过期
	Voted                   //本Term内已经投过票
)

//TODO
const (
	AppNormal    AppendEntriesState = iota // 追加正常
	AppOutOfDate                           // 追加过时
	AppKilled                              // Raft程序终止
	AppRepeat                              // 追加重复 (2B
	AppCommited                            // 追加的日志已经提交 (2B
	Mismatch                               // 追加不匹配 (2B
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

const (

	// MoreVoteTime MinVoteTime 定义随机生成投票过期时间范围:(MoreVoteTime+MinVoteTime~MinVoteTime)
	MoreVoteTime = 100
	MinVoteTime  = 75

	// HeartbeatSleep 心脏休眠时间,要注意的是，这个时间要比选举低，才能建立稳定心跳机制
	HeartbeatSleep = 35
	AppliedSleep   = 15
)

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu    sync.Mutex          // Lock to protect shared access to this peer's state
	peers []*labrpc.ClientEnd // RPC end points of all peers

	//类似于持久化相关的内容
	persister *Persister // Object to hold this peer's persisted state
	me        int        // this peer's index into peers[]
	dead      int32      // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	// 所有的servers需要持久化的变量:
	currentTerm int        // 记录当前的任期
	votedFor    int        // 记录当前的任期把票投给了谁
	logs        []LogEntry // 日志条目数组，包含了状态机要执行的指令集，以及收到领导时的任期号

	// 对于正常流程来说应该先applyLog进chan里，然后更新commit，最后两者应该是相同的，只是先后更新顺序不同
	commitIndex int
	lastApplied int

	// nextIndex与matchIndex初始化长度应该为len(peers)，Leader对于每个Follower都记录他的nextIndex和matchIndex
	// nextIndex指的是下一个的appendEntries要从哪里开始
	// matchIndex指的是已知的某follower的log与leader的log最大匹配到第几个Index,已经apply
	nextIndex  []int // 对于每一个server，需要发送给他下一个日志条目的索引值（初始化为leader日志index+1,那么范围就对标len）
	matchIndex []int // 对于每一个server，已经复制给该server的最后日志条目下标

	applyChan chan ApplyMsg // 用来写入通道

	// paper外自己追加的
	status     Status
	voteNum    int // 记录当前投票给了谁
	votedTimer time.Time

	// 2D中用于传入快照点
	lastIncludeIndex int
	lastIncludeTerm  int
}

type LogEntry struct {
	Term    int
	Command interface{}
}

// -------------------------RPC参数部分----------------------------
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int //	需要竞选的人的任期
	CandidateId  int // 需要竞选的人的Id
	LastLogIndex int // 竞选人日志条目最后索引(2D包含快照
	LastLogTerm  int // 候选人最后日志条目的任期号(2D包含快照
}

// RequestVoteReply
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// Your data here (2A).
	Term        int  // 投票方的term，如果竞选者比自己还低就改为这个
	VoteGranted bool // 是否投票给了该竞选人
}

// AppendEntriesArgs Append Entries RPC structure
type AppendEntriesArgs struct {
	Term         int        // leader的任期
	LeaderId     int        // leader自身的ID
	PrevLogIndex int        // 用于匹配日志位置是否是合适的，初始化rf.nextIndex[i] - 1
	PrevLogTerm  int        // 用于匹配日志的任期是否是合适的是，是否有冲突
	Entries      []LogEntry // 预计存储的日志（为空时就是心跳连接）
	LeaderCommit int        // leader的commit index指的是最后一个被大多数机器都复制的日志Index
}

type AppendEntriesReply struct {
	Term        int  // leader的term可能是过时的，此时收到的Term用于更新他自己
	Success     bool //	如果follower与Args中的PreLogIndex/PreLogTerm都匹配才会接过去新的日志（追加），不匹配直接返回false
	UpNextIndex int  // 如果发生conflict时reply传过来的正确的下标用于更新nextIndex[i]
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	rf.mu.Lock()
	defer rf.mu.Unlock()

	return rf.currentTerm, rf.status == Leader
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
//TODO what it is
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
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
// within a timeout interval, Call() returns true; otherwise
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
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {

	// Your code here (2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.killed() == true {
		return -1, -1, false
	}
	if rf.status != Leader {
		return -1, -1, false
	} else {
		index := rf.getLastIndex() + 1
		term := rf.currentTerm
		rf.logs = append(rf.logs, LogEntry{Term: term, Command: command})
		rf.persist()
		return index, term, true
	}
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

// The ticker go routine starts a new election if this peer hasn't received
// heartsbeats recently.
func (rf *Raft) ticker() {
	for rf.killed() == false {

		// Your code here to check if a leader election should
		// be started and to randomize sleeping time using
		// time.Sleep().

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
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).
	rf.mu.Lock()

	rf.status = Follower
	rf.currentTerm = 0
	rf.voteNum = 0
	rf.votedFor = -1

	rf.lastApplied = 0
	rf.commitIndex = 0

	rf.lastIncludeIndex = 0
	rf.lastIncludeTerm = 0

	rf.logs = []LogEntry{}
	rf.logs = append(rf.logs, LogEntry{})
	rf.applyChan = applyCh
	rf.mu.Unlock()

	// initialize from status persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// 同步快照信息
	if rf.lastIncludeIndex > 0 {
		rf.lastApplied = rf.lastIncludeIndex
	}

	go rf.electionTicker()
	//
	//go rf.appendTicker()
	//
	//go rf.committedTicker()

	return rf
}

// ---------------------心跳部分-----------------------------------
func (rf *Raft) electionTicker() {
	for rf.killed() == false {
		nowTime := time.Now()
		time.Sleep(time.Duration(generateOverTime(int64(rf.me))) * time.Millisecond)
		rf.mu.Lock()

		// 时间过期发起选举
		// 此处的流程为每次每次votedTimer如果小于在sleep睡眠之前定义的时间，就代表没有votedTimer没被更新为最新的时间，则发起选举
		if rf.votedTimer.Before(nowTime) && rf.status != Leader {
			// 转变状态
			rf.status = Candidate
			rf.votedFor = rf.me
			rf.voteNum = 1
			rf.currentTerm += 1
			rf.persist()

			//fmt.Printf("[++++elect++++] :Rf[%v] send a election\n", rf.me)
			rf.sendElection()
			rf.votedTimer = time.Now()
		}
		rf.mu.Unlock()
	}
}

func (rf *Raft) appendTicker() {
	for rf.killed() == false {
		time.Sleep(HeartbeatSleep * time.Millisecond)
		rf.mu.Lock()
		if rf.status == Leader {
			rf.mu.Unlock()
			rf.leaderAppendEntries()
		} else {
			rf.mu.Unlock()
		}
	}
}

//---------------------------------------leader选举部分-------------------------------------

func (rf *Raft) sendElection() {
	for i := 0; i < len(rf.peers); i++ {
		if i == rf.me {
			continue
		}

		// 开启协程对各个节点发起选举
		go func(server int) {
			rf.mu.Lock()
			args := RequestVoteArgs{
				rf.currentTerm,
				rf.me,
				rf.getLastIndex(),
				rf.getLastTerm(),
			}
			reply := RequestVoteReply{}
			rf.mu.Unlock()
			res := rf.sendRequestVote(server, &args, &reply)

			if res == true {
				rf.mu.Lock()
				// 判断自身是否还是竞选者，且任期不冲突
				if rf.status != Candidate || args.Term < rf.currentTerm {
					rf.mu.Unlock()
					return
				}

				// 返回者的任期大于args（网络分区原因)进行返回
				if reply.Term > args.Term {
					if rf.currentTerm < reply.Term {
						rf.currentTerm = reply.Term
					}
					rf.status = Follower
					rf.votedFor = -1
					rf.voteNum = 0
					//TODO what is it
					rf.persist()
					rf.mu.Unlock()
					return
				}

				// 返回结果正确判断是否大于一半节点同意
				if reply.VoteGranted == true && rf.currentTerm == args.Term {
					rf.voteNum += 1
					if rf.voteNum >= len(rf.peers)/2+1 {

						//fmt.Printf("[++++elect++++] :Rf[%v] to be leader,term is : %v\n", rf.me, rf.currentTerm)
						rf.status = Leader
						rf.votedFor = -1
						rf.voteNum = 0
						rf.persist()

						//TODO：question 这一块的意义
						rf.nextIndex = make([]int, len(rf.peers))
						for i := 0; i < len(rf.peers); i++ {
							rf.nextIndex[i] = rf.getLastIndex() + 1
						}
						rf.matchIndex = make([]int, len(rf.peers))
						rf.matchIndex[rf.me] = rf.getLastIndex()

						rf.votedTimer = time.Now()
						rf.mu.Unlock()
						return
					}
					rf.mu.Unlock()
					return
				}

				rf.mu.Unlock()
				return
			}

		}(i)

	}
}

// RequestVote
// example RequestVote RPC handler.
// 个人认为定时刷新的地方应该是别的节点与当前节点在数据上不冲突时才要刷新
// 因为如果不是数据冲突那么定时相当于防止自身去选举的一个心跳
// 如果是因为数据冲突，那么这个节点不用刷新定时是为了当前整个raft能尽快有个正确的leader
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// 由于网络分区或者是节点crash，导致的任期比接收者还小，直接返回
	if args.Term < rf.currentTerm {
		reply.VoteGranted = false
		reply.Term = rf.currentTerm
		return
	}

	reply.Term = rf.currentTerm

	// 预期的结果:任期大于当前节点，进行重置
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.status = Follower
		rf.votedFor = -1
		rf.voteNum = 0
		rf.persist()
	}

	// If votedFor is null or candidateId, and candidate’s logs is at
	// least as up-to-date as receiver’s logs, grant vote
	//TODO Remove rf.votedFor != args.CandidateId && args.Term == reply.Term
	if !rf.UpToDate(args.LastLogIndex, args.LastLogTerm) || // 判断日志是否conflict
		rf.votedFor != -1 && rf.votedFor != args.CandidateId && args.Term == rf.currentTerm { // paper中的第二个条件votedFor is null

		// 满足以上两个其中一个都返回false，不给予投票
		reply.VoteGranted = false
		reply.Term = rf.currentTerm
		return
	} else {
		reply.VoteGranted = true
		rf.votedFor = args.CandidateId
		rf.currentTerm = args.Term
		reply.Term = rf.currentTerm
		rf.votedTimer = time.Now()
		rf.persist()
		return
	}

}

//---------------------------------------日志增量部分------------------------------------------
func (rf *Raft) leaderAppendEntries() {}
