package ipc

/*
#include <sys/ipc.h>
#include <sys/sem.h>
#include <sched.h>
#include <stdlib.h>
#include <errno.h>

int get_sem_stat(int semid, struct semid_ds *buf) {
	return semctl(semid, 0, IPC_STAT, buf);
}
int init_sem(int semid, int val) {
	if (semctl(semid, 0, SETVAL, val-1) == -1) return -1;
	struct sembuf sop;
	sop.sem_num = 0;
	sop.sem_op = 1;
	sop.sem_flg = 0;
	return semop(semid, &sop, 1);
}
int remove_sem(int semid) {
	return semctl(semid, 0, IPC_RMID);
}
int op_sem(int semid, int op, int flag) {
	struct sembuf sop;
	sop.sem_num = 0;
	sop.sem_op = op;
	sop.sem_flg = flag;
	return semop(semid, &sop, 1);
}
int get_sem(int semid) {
	unsigned short i;
	semctl(semid, 0, GETALL, &i);
	return i;
}
void try_wait_sem_undo(int semid) {
	op_sem(semid, -1, SEM_UNDO | IPC_NOWAIT);
}
void wait_sem_undo(int semid) {
	op_sem(semid, -1, SEM_UNDO);
}
void signal_sem_undo(int semid) {
	op_sem(semid, 1, SEM_UNDO);
}
void try_wait_sem(int semid) {
	op_sem(semid, -1, IPC_NOWAIT);
}
void wait_sem(int semid) {
	op_sem(semid, -1, 0);
}
void signal_sem(int semid) {
	op_sem(semid, 1, 0);
}
*/
import "C"

import (
	"log"
	"syscall"
	"time"
	"unsafe"
)

const (
	BusyWaitTimeout = time.Millisecond * 500
)

func GenFileKey(pathName string, projectId int) C.key_t {
	c := C.CString(pathName)
	defer C.free(unsafe.Pointer(c))
	r, err := C.ftok(c, C.int(projectId))
	if err != nil {
		log.Panicln("ftok:", pathName, projectId, err)
	}
	return r
}

type Semaphore C.int

func NewSemaphore(pathName string, projectId int, initVal int) Semaphore {
	semKey := GenFileKey(pathName, projectId)
	semId, err := C.semget(semKey, 1, 0600|C.IPC_CREAT|C.IPC_EXCL)
	if err == nil {
		_, err = C.init_sem(semId, C.int(initVal))
		if err != nil {
			log.Fatalln(err)
		}
	} else if err == syscall.EEXIST {
		semId, err = C.semget(semKey, 1, 0600)
		if err != nil {
			log.Fatalln(err)
		}
		var sds C.struct_semid_ds
		var timeout = time.After(BusyWaitTimeout)
		for sds.sem_otime == 0 {
			select {
			case <-timeout:
				_, err = C.init_sem(semId, C.int(initVal))
				log.Println("busywait: timeout", pathName, projectId)
				if err != nil {
					log.Fatalln(err)
				}
			default:
				C.sched_yield()
				_, err = C.get_sem_stat(semId, &sds)
				if err != nil {
					log.Fatalln(err)
				}
			}
		}
	} else {
		log.Fatalln(err)
	}
	return Semaphore(semId)
}
func (s Semaphore) Wait()   { C.wait_sem_undo(C.int(s)) }
func (s Semaphore) Signal() { C.signal_sem_undo(C.int(s)) }
func (s Semaphore) TryWait() bool {
	_, err := C.try_wait_sem_undo(C.int(s))
	return err == nil
}
func (s Semaphore) WaitHold()   { C.wait_sem(C.int(s)) }
func (s Semaphore) SignalHold() { C.signal_sem(C.int(s)) }
func (s Semaphore) TryWaitHold() bool {
	_, err := C.try_wait_sem(C.int(s))
	return err == nil
}
func (s Semaphore) Remove() error {
	_, err := C.remove_sem(C.int(s))
	return err
}

type Mutex struct {
	s    Semaphore
	asym bool
}

func NewMutex(pathName string, projectId int, symmetric bool) Mutex {
	return Mutex{NewSemaphore(pathName, projectId, 1), !symmetric}
}
func (m Mutex) Lock() {
	if m.asym {
		m.s.WaitHold()
	} else {
		m.s.Wait()
	}
}
func (m Mutex) Unlock() {
	if m.asym {
		m.s.SignalHold()
	} else {
		m.s.Signal()
	}
}
func (m Mutex) TryLock() bool {
	if m.asym {
		return m.s.TryWaitHold()
	} else {
		return m.s.TryWait()
	}
}
func (m Mutex) Remove() error {
	return m.s.Remove()
}
