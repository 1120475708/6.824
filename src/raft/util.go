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

func (rf *Raft) getLastIndex() int {
	return len(rf.logs) - 1 + rf.lastIncludeIndex
}

func (rf *Raft) restoreLog(curIndex int) LogEntry {
	return rf.logs[curIndex-rf.lastIncludeIndex]
}
