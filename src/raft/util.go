package raft

import (
	"log"
	"math/rand"
	"time"
)

// Debugging
const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

func generateOverTime(server int64) int {
	rand.Seed(time.Now().Unix() + server)
	return rand.Intn(MoreVoteTime) + MinVoteTime
}

// 获取最后的快照日志下标(代表已存储）
func (rf *Raft) getLastIndex() int {
	return len(rf.logs) - 1 + rf.lastIncludeIndex
}

// 获取最后的任期(快照版本
func (rf *Raft) getLastTerm() int {
	// 因为初始有填充一个，否则最直接len == 0
	if len(rf.logs)-1 == 0 {
		return rf.lastIncludeTerm
	} else {
		return rf.logs[len(rf.logs)-1].Term
	}
}

// 通过快照偏移还原真实日志条目
func (rf *Raft) restoreLog(curIndex int) LogEntry {
	return rf.logs[curIndex-rf.lastIncludeIndex]
}

// UpToDate paper中投票RPC的rule2
func (rf *Raft) UpToDate(index int, term int) bool {
	lastIndex := rf.getLastIndex()
	lastTerm := rf.getLastTerm()
	//leader任期大于当前任期，leader任期等于当前任期，并且leader日志index大于当前索引
	return term > lastTerm || (term == lastTerm && index >= lastIndex)
}
