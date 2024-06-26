package main

import (
	"bufio"
	"container/list"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// globals
var EMPTYPCB int = 0
var DELETEDPROCESSCOUNTER int = 0
var first int = 0

type pcb struct {
	State     int
	Parent    int
	Priority  int //1 to n -1
	Children  *list.List
	Resources *list.List
}

type rcb struct {
	State     int
	Inventory int
	Waitlist  *list.List
}

type waitlistProcess struct {
	ProcessIndex   int //index of the process in the pcb
	requestedUnits int
}

// pcb resources is a linked list of these
type resourceInfo struct {
	ResourceIndex int
	Units         int
}

type Command struct {
	Type string
	Args []string
}

// finds the element inside of a linked list regardless of type of linked list
func findElement(index int, list *list.List) interface{} {
	var counter int = 0
	for e := list.Front(); e != nil; e = e.Next() {
		if counter == index {
			return e.Value
		}
		counter++
	}
	return nil
}

// find the empty slot (index) in pcb to add a new process pcb
func findEmptyPCB(pcbArray []*pcb) int {
	for e := 0; e < len(pcbArray); e++ {
		if pcbArray[e] == nil {
			return e
		}

	}
	return -1
}

// find the empty bucket in an array (mainly used for ready list)
func findEmptySlot(arr []int) int {
	for i := 0; i < len(arr); i++ {
		if arr[i] == -1 {
			return i
		}
	}
	return -1
}

// find the running process based on the head of the ready list, search top down from 2 (highest)
// RETURN: index of the running process in the pcb array
func findRunningProcess(readyList *[][]int) int {
	n := len(*readyList)

	for i := n - 1; i >= 0; i-- {
		//the running process will always be the first element in the inner arrays
		if (*readyList)[i][0] != -1 {
			return (*readyList)[i][0]
		}
	}
	//if all of them are empty, this is the beginning of init
	return -1
}

// finds the running process in ready list which iterates through each level and finds the level with a non null array
// removes the first element in that level
func readyListRemoval(readyList *[][]int) {
	var n int = len(*readyList)

	//from n to 1 bc 0 doesnt have a running reserved for init
	for i := n - 1; i > 0; i-- {
		//non null means that the running process is there
		if (*readyList)[i][0] != -1 {
			//shifts over the head to remove
			for j := 1; j < 16; j++ {
				(*readyList)[i][j-1] = (*readyList)[i][j]
			}
			break
		}
	}
}

func resourceListRemoval(pcbArray *[]*pcb, pcbIndex int, resourceIndex int) {
	for e := (*pcbArray)[pcbIndex].Resources.Front(); e != nil; e = e.Next() {
		//find the node in the resources linked list and remove
		frontResourceIndex := e.Value.(*resourceInfo).ResourceIndex
		if frontResourceIndex == resourceIndex {
			(*pcbArray)[pcbIndex].Resources.Remove(e)
			//eString := strconv.Itoa(frontResourceIndex)
			//fmt.Printf("Removing resource index: " + eString)
			break
		}

	}
}

func waitListRemoval(rcbArray *[]*rcb, resourceIndex int, requestedUnits int, pcbArray *[]*pcb, readyList *[][]int) {
	for {
		// while (r.waitlist != empty and r.state > 0)
		if ((*rcbArray)[resourceIndex].Waitlist.Len() == 0) || ((*rcbArray)[resourceIndex].State <= 0) {
			break
		}

		// get next (j, k) from r.waitlist, j is the process index and k is the amt of resources requested by process
		unblockedProcess, ok := (*rcbArray)[resourceIndex].Waitlist.Front().Value.(*waitlistProcess)
		if !ok {
			panic(ok)
		}
		// if (r.state >= k)
		if (*rcbArray)[resourceIndex].State >= unblockedProcess.requestedUnits {
			// r.state = r.state - k
			(*rcbArray)[resourceIndex].State = (*rcbArray)[resourceIndex].State - unblockedProcess.requestedUnits
			// insert (r, k) into j.resources
			newResource := &resourceInfo{ResourceIndex: resourceIndex, Units: requestedUnits}
			(*pcbArray)[unblockedProcess.ProcessIndex].Resources.PushBack(newResource)
			// j.state = ready
			(*pcbArray)[unblockedProcess.ProcessIndex].State = 1
			// remove (j, k) from r.waitlist
			//var removed bool = false
			for e := (*rcbArray)[resourceIndex].Waitlist.Front(); e != nil; e = e.Next() {
				node := e.Value.(*waitlistProcess)
				if unblockedProcess == node {
					(*rcbArray)[resourceIndex].Waitlist.Remove(e)
					//removed = true
				}
			}

			//insert j into RL
			var newPrio int = (*pcbArray)[unblockedProcess.ProcessIndex].Priority

			emptySlot := findEmptySlot((*readyList)[newPrio])
			(*readyList)[newPrio][emptySlot] = unblockedProcess.ProcessIndex
		} else {
			break
		}

	}

}

// iterates through the children list of parent to see if the child exists
func checkChild(pcbArray *[]*pcb, parent int, child int) bool {
	childrenList := (*pcbArray)[parent].Children

	for e := childrenList.Front(); e != nil; e = e.Next() {
		if e.Value.(int) == child {
			return true
		}
	}
	return false
}

func releaseEverything(pcbArray *[]*pcb, rcbArray *[]*rcb, readyList *[][]int, childInt int) {
	//remove j from parent’s list of children
	//iterate through running process children list and remove the e
	//for e := (*pcbArray)[currIndex].Children.Front(); e != nil; e = e.Next() {
	//	if e.Value.(int) == childInt {
	//		(*pcbArray)[currIndex].Children.Remove(e)
	//	}
	//}

	rlFlag := false
	//remove j from RL
	//check if j is in RL, if so remove
	for x := (len(*readyList) - 1); x >= 0; x-- {
		if rlFlag {
			break
		}
		//check each innerArray for the child int
		for y := 0; y < 16; y++ {
			if (*readyList)[x][y] == childInt {
				//if it's in the front -> shift
				if y == 0 {
					for z := 1; z < 16; z++ {
						(*readyList)[x][z-1] = (*readyList)[x][z]
					}
				} else if y > 0 && y < 15 { //middle -> shift right of middle to the left
					for z := y + 1; z < 16; z++ {
						(*readyList)[x][z-1] = (*readyList)[x][z]
					}
				} else if (*readyList)[x][y+1] == -1 || y == 15 { //end -> assign  -1 to it
					(*readyList)[x][y] = -1
				}
				//fmt.Printf("Child removed from RL")
				rlFlag = true
				break
			}
		}
	}

	wlFlag := false
	//remove j from WL if exists
	//iterate through the RCB Array and search
	for x := 0; x < 4; x++ {
		if wlFlag {
			break
		}
		for e := (*rcbArray)[x].Waitlist.Front(); e != nil; e = e.Next() {
			actualValue := e.Value.(*waitlistProcess)
			if actualValue.ProcessIndex == childInt {
				(*rcbArray)[x].Waitlist.Remove(e)
				//fmt.Printf("Child removed from WL")
				wlFlag = true
				break
			}
		}
	}
	//check if j is holding a resource and if so remove from the rcb
	if (*pcbArray)[childInt].Resources.Len() != 0 {
		//iterate through the resources and get the resource index
		//and add back on the units held into the rcb state for that certain index
		for x := (*pcbArray)[childInt].Resources.Front(); x != nil; x = x.Next() {
			rcbIndex := x.Value.(*resourceInfo).ResourceIndex
			(*rcbArray)[rcbIndex].State += x.Value.(*resourceInfo).Units
			//update the waitlist
			waitListRemoval(rcbArray, rcbIndex, x.Value.(*resourceInfo).Units, pcbArray, readyList)
		}
	}

	//release all resources of j
	(*pcbArray)[childInt].Resources = list.New()

	//free PCB of j, and index of pcb can never be reused
	(*pcbArray)[childInt] = nil

}

// starts off with j's child
// recursively go through all children and release everything
func descendantDeletion(children *list.List, pcbArray *[]*pcb, rcbArray *[]*rcb, readyList *[][]int) {
	//base case: children list is nil return
	if children == nil {
		return
	}
	//iterate through the children list and recursively call descendant deletion
	for e := children.Front(); e != nil; {
		childIndex := e.Value.(int)
		next := e.Next()
		descendantDeletion((*pcbArray)[childIndex].Children, pcbArray, rcbArray, readyList)
		releaseEverything(pcbArray, rcbArray, readyList, childIndex)

		//after releasing the child int from all the data structures, note the rcbArray indexes that were released
		//seperate function to check what you can remove from the waitlist

		children.Remove(e)
		e = next
		DELETEDPROCESSCOUNTER++
	}
	return
}

func scheduler(readyList [][]int) {
	var n int = len(readyList)

	//from n to 1 bc 0 doesnt have a running reserved for init
	for i := n - 1; i > 0; i-- {
		//non null means that the running process is there
		if (readyList)[i][0] != -1 {
			head := strconv.Itoa(readyList[i][0])
			//fmt.Printf("Process: " + head + " running")
			fmt.Printf("%s ", head)
			return
		}
	}

	//fmt.Printf("Process: 0 running")
	fmt.Printf("0 ")
}

// Init: n = amt of priority levels | u_num = the amt of units for resource_num
// Notes: creates ready list with n priority levels 0 to n -1, and returns it
func in(n, u0, u1, u2, u3 string, pcbArray *[]*pcb, rcbArray *[]*rcb) [][]int {
	//var cmdLine string = fmt.Sprintf("p1: %s, p2: %s, p3: %s, p4: %s, p5: %s", n, u0, u1, u2, u3)
	//fmt.Printf("all of the inputs for in: " + cmdLine)

	//to print a line before each output after the first one
	if first > 0 {
		fmt.Printf("\n")
	}
	first++

	var int0, _ = strconv.Atoi(u0)
	var int1, _ = strconv.Atoi(u1)
	var int2, _ = strconv.Atoi(u2)
	var int3, _ = strconv.Atoi(u3)

	prioLevels, _ := strconv.Atoi(n)

	//clear everything back to default
	EMPTYPCB = 0

	(*pcbArray) = make([]*pcb, 16)

	for i := range *pcbArray {
		(*pcbArray)[i] = nil
	}

	(*rcbArray) = make([]*rcb, 4)

	for i := range *rcbArray {
		(*rcbArray)[i] = nil
	}

	if prioLevels <= 0 {
		//fmt.Printf("ERROR: must have at least 1 priority level")
		fmt.Printf("-1 ")
		return nil
	} else {
		//initializes the rcb array with the params
		var rcb0 rcb = rcb{
			State:     int0,
			Inventory: int0,
			Waitlist:  list.New(),
		}

		var rcb1 rcb = rcb{
			State:     int1,
			Inventory: int1,
			Waitlist:  list.New(),
		}

		var rcb2 rcb = rcb{
			State:     int2,
			Inventory: int2,
			Waitlist:  list.New(),
		}

		var rcb3 rcb = rcb{
			State:     int3,
			Inventory: int3,
			Waitlist:  list.New(),
		}

		//lenString := strconv.Itoa(len(*rcbArray))
		//fmt.Printf("size of rcb array: " + lenString)
		(*rcbArray)[0] = &rcb0
		(*rcbArray)[1] = &rcb1
		(*rcbArray)[2] = &rcb2
		(*rcbArray)[3] = &rcb3

		//initializes a 2d ready list of n buckets with len 16 in each bucket
		readyList := make([][]int, prioLevels)

		for i := 0; i < prioLevels; i++ {
			innerArray := make([]int, 16)
			for j := 0; j < 16; j++ {
				innerArray[j] = -1
			}
			readyList[i] = innerArray
		}

		//intializes the pcbArray
		create(&readyList, pcbArray, "0")
		EMPTYPCB++
		//fmt.Printf("Successfully initialized!")
		//scheduler(readyList)
		return readyList
	}
}

// Create: p = priority level (1,2,0 but 0 is for init process)
func create(readyList *[][]int, pcbArray *[]*pcb, p string) {
	//allocate new PCB[j]
	//getes the empty slot to insert the new process pcb
	runningIndex := findRunningProcess(readyList)
	priority, _ := strconv.Atoi(p)

	//if there is no running process in ready list, this is init calling create
	if runningIndex == -1 {
		//creates the init pcb running with nothing at prio level 0
		var newPCB pcb = pcb{
			State:     -1,
			Parent:    -1,
			Priority:  0,
			Children:  list.New(),
			Resources: list.New(),
		}

		//add to the pcbArray and to the readyList
		(*pcbArray)[0] = &newPCB

		//ready list prio 0 at the head is index 0 of the init pcb
		(*readyList)[0][0] = 0
	} else {
		//if there is a running process it is the one that calls create
		//assign the new pcb to the running process's child and vice versa new pcb parent = running
		if EMPTYPCB == 16 {
			//fmt.Printf("ERROR: empty slot is  -1, too many processes")
			fmt.Printf("-1 ")
			return
		} else { //running process creates a child
			var newPCB pcb = pcb{
				State:     1,
				Parent:    runningIndex,
				Priority:  priority,
				Children:  list.New(),
				Resources: list.New(),
			}

			//updating the running process children list
			(*pcbArray)[runningIndex].Children.PushBack(EMPTYPCB)

			//add the new process to pcb array
			(*pcbArray)[EMPTYPCB] = &newPCB

			if priority == 0 {
				//fmt.Printf("ERROR: not init -> cannot add process in priority level 0")
				fmt.Printf("-1 ")
				return
			}

			emptySlot := findEmptySlot((*readyList)[priority])

			//add to readylist
			(*readyList)[priority][emptySlot] = EMPTYPCB

			//fmt.Printf("Process: " + strconv.Itoa(EMPTYPCB) + " created successfully!")
		}
	}

}

// Destroy: i = pcb index
func destroy(pcbArray *[]*pcb, rcbArray *[]*rcb, readyList *[][]int, j string) int {
	runningIndex := findRunningProcess(readyList)
	childInt, _ := strconv.Atoi(j)
	//check if j is a child process of the running process
	if !checkChild(pcbArray, runningIndex, childInt) {
		//TODO: TEST DELETING ITSELF
		//check if j == i, it's trying to destroy itself
		if runningIndex == childInt {
			//destroy all of itself's descendants
			descendantDeletion((*pcbArray)[childInt].Children, pcbArray, rcbArray, readyList)
			//release itself from everything
			releaseEverything(pcbArray, rcbArray, readyList, childInt)
		}

		//fmt.Printf("DESTROY ERROR: j: " + j + " doesn't exist in the running process")
		return -1
	}
	//initialize back to
	DELETEDPROCESSCOUNTER = 0
	//for all k in children of j : destroy(k)
	//recursively destroy j and it's descendants
	//pass in the head of the children list and keep going til there's an empty list and return
	//after clearing all the children make it a new list
	descendantDeletion((*pcbArray)[childInt].Children, pcbArray, rcbArray, readyList)

	(*pcbArray)[childInt].Children = list.New()

	//remove j from parent’s list of children
	//iterate through running process children list and remove the e
	for e := (*pcbArray)[runningIndex].Children.Front(); e != nil; e = e.Next() {
		if e.Value.(int) == childInt {
			(*pcbArray)[runningIndex].Children.Remove(e)
		}
	}

	releaseEverything(pcbArray, rcbArray, readyList, childInt)
	////display: “n processes destroyed”
	//fmt.Printf(counterString + " process recursively destroyed")

	return 1
}

// Request: r = resource number | k = num of units for resource r
func request(readyList *[][]int, pcbArray *[]*pcb, rcbArray *[]*rcb, r string, k string) int {
	runningIndex := findRunningProcess(readyList)
	resourceNum, _ := strconv.Atoi(r)
	requestedUnits, _ := strconv.Atoi(k)
	inventory := (*rcbArray)[resourceNum].Inventory
	state := (*rcbArray)[resourceNum].State

	if requestedUnits <= 0 {
		//fmt.Printf("ERROR: the amount of units requested has to be greater than 0")
		return -1
	}

	// if state of r is free
	//num of units requested + num alr held <= initial inventory
	//k + (inventory - state) <= inventory
	//fails check then automatically return  -1
	if requestedUnits+(inventory-state) <= inventory {
		// state of r = allocated
		(*rcbArray)[resourceNum].State = state - requestedUnits
		// insert r into list of resources of process i
		newResource := &resourceInfo{ResourceIndex: resourceNum, Units: requestedUnits}
		(*pcbArray)[runningIndex].Resources.PushBack(newResource)
		// display: “resource r allocated”
		//fmt.Printf("Resource: " + r + " allocated!")

		return 1 //1 is successfully allocated
	} else {
		// state of i = blocked -> state in pcb becomes 0
		(*pcbArray)[runningIndex].State = 0
		//remove i (head of RL) from RL
		readyListRemoval(readyList)
		//add (i,k) to waitlist of r
		newWaitlist := &waitlistProcess{ProcessIndex: runningIndex, requestedUnits: requestedUnits}
		(*rcbArray)[resourceNum].Waitlist.PushBack(newWaitlist)
		// display: “process i blocked”
		//runningIndexString := strconv.Itoa(runningIndex)
		//fmt.Printf("Process: " + runningIndexString + " blocked!")

		return 0 //0 is blocked
	}

}

// Release: r = resource number | k = num of units for resource r
func release(readyList *[][]int, pcbArray *[]*pcb, rcbArray *[]*rcb, r string, k string) int {
	runningIndex := findRunningProcess(readyList)
	resourceNum, _ := strconv.Atoi(r)
	requestedUnits, _ := strconv.Atoi(k)
	inventory := (*rcbArray)[resourceNum].Inventory
	state := (*rcbArray)[resourceNum].State

	//error check: num of units <= num of units currently held
	//k <= (inventory - state)
	if requestedUnits > (inventory - state) {
		//fmt.Printf("ERROR: released more than the number of units currently held for resource: " + r)
		return -1
	}

	// remove r from resources list of process i
	resourceListRemoval(pcbArray, runningIndex, resourceNum)

	(*rcbArray)[resourceNum].State = state + requestedUnits
	waitListRemoval(rcbArray, resourceNum, requestedUnits, pcbArray, readyList)

	//// if waitlist of r is empty
	//if (*rcbArray)[resourceNum].Waitlist == nil {
	//	// state of r = free (state goes back to original amt of units)
	//} else {
	//	//remove process j (head of WL) from the WL bc no longer blocked by process i -> joins RL
	//}

	return 1
}

// Timeout
func timeout(readyList *[][]int) {
	//store the og running process
	var runningProcess int = findRunningProcess(readyList)
	//iterate from n -1 to 1, and find the first array that isn't nil
	var runningPriority int
	for i := (len(*readyList) - 1); i > 0; i-- {
		if (*readyList)[i][0] != -1 {
			runningPriority = i
			break
		}
	}
	//remove the running process and append to the back on the same prio level same array
	readyListRemoval(readyList)
	emptySlot := findEmptySlot((*readyList)[runningPriority])

	//the left most empty bucket in the inner array
	(*readyList)[runningPriority][emptySlot] = runningProcess

}

func runCommand(pcbArray *[]*pcb, rcbArray *[]*rcb, rl *[][]int, c Command) {
	cmd := c.Type
	//fmt.Printf("type is " + cmd)
	input := c.Args
	var p1, p2, p3, p4, p5 string

	switch cmd {

	case "in":
		////fmt.Printf("Command: " + cmd)
		if len(input) == 5 {
			p1 = input[0]
			p2 = input[1]
			p3 = input[2]
			p4 = input[3]
			p5 = input[4]

			*rl = in(p1, p2, p3, p4, p5, pcbArray, rcbArray)
			scheduler(*rl)
		} else {
			fmt.Printf("Not enough params for the cmd: in")
		}
	case "cr":
		//fmt.Printf("Command: " + cmd)
		if rl == nil {
			//fmt.Printf("ERROR: have to run in/id first")
			break
		}
		if len(input) == 1 {
			p1 = input[0]
			priorityInt, _ := strconv.Atoi(p1)
			if priorityInt < 0 || priorityInt >= len(*rl) {
				//fmt.Printf("ERROR_CR: the priority number is out of range")
				fmt.Printf("-1 ")
				break
			}
			create(rl, pcbArray, p1)
			scheduler(*rl)

			EMPTYPCB++
		} else {
			//fmt.Printf("Not enough params for the cmd: cr")
			fmt.Printf("-1 ")
		}
	case "de":
		//fmt.Printf("Command: " + cmd)
		if rl == nil {
			//fmt.Printf("ERROR: have to run in/id first")
			fmt.Printf("-1 ")
			break
		}

		if len(input) == 1 {
			p1 = input[0]
			if p1 == "0" {
				fmt.Printf("-1 ")
			}
			if destroy(pcbArray, rcbArray, rl, p1) == -1 {
				//fmt.Printf("Destroy Error")
				fmt.Printf("-1 ")

			} else {
				scheduler(*rl)
			}
		} else {
			//fmt.Printf("Not enough params for the cmd: de")
			fmt.Printf("-1 ")
		}
	case "rq":
		//fmt.Printf("Command: " + cmd)
		if rl == nil {
			//fmt.Printf("ERROR: have to run in/id first")
			break
		}

		if len(input) == 2 {
			p1 = input[0]
			p2 = input[1]

			resourceNum, _ := strconv.Atoi(p1)

			if resourceNum < 0 || resourceNum > 3 {
				//fmt.Printf("ERROR_RQ: resource num not in the right range")
				fmt.Printf("-1 ")
				break
			}

			if request(rl, pcbArray, rcbArray, p1, p2) == -1 {
				//fmt.Printf("ERROR: failed to request")
				fmt.Printf("-1 ")
				break
			} else {
				scheduler(*rl)
			}
		} else {
			//fmt.Printf("Not enough params for the cmd: rq")
			fmt.Printf("-1 ")
		}

	case "rl":
		//fmt.Printf("Command: " + cmd)
		if rl == nil {
			//fmt.Printf("ERROR: have to run in/id first")
			fmt.Printf("-1 ")
			break
		}

		if len(input) == 2 {
			p1 = input[0]
			p2 = input[1]

			resourceNum, _ := strconv.Atoi(p1)
			requestedUnits, _ := strconv.Atoi(p2)

			if resourceNum < 0 || resourceNum > 3 {
				fmt.Printf("-1 ")
				break
			}

			inventory := (*rcbArray)[resourceNum].Inventory

			//state := (*rcbArray)[resourceNum].State

			if requestedUnits < 0 || requestedUnits > inventory {
				//fmt.Printf("Process  -1 running")
				fmt.Printf("-1 ")

				break

			}

			if release(rl, pcbArray, rcbArray, p1, p2) == -1 {
				//fmt.Printf("ERROR: failed to release")
				//fmt.Printf("Process  -1 running")
				fmt.Printf("-1 ")

			} else {
				scheduler(*rl)
			}
		} else {
			//fmt.Printf("Not enough params for the cmd: rl")
			fmt.Printf("-1 ")

		}
	case "to":
		//fmt.Printf("Command: " + cmd)
		if rl == nil {
			//fmt.Printf("ERROR: have to run in/id first")
			fmt.Printf("-1 ")

			break
		}

		if len(input) == 0 {
			timeout(rl)
			scheduler(*rl)
		} else {
			//fmt.Printf("Not enough params for the cmd: rq")
			fmt.Printf("-1 ")

		}
	case "id":
		//fmt.Printf("Command: " + cmd)
		if len(input) == 0 {

			*rl = in("3", "1", "1", "2", "3", pcbArray, rcbArray)

			scheduler(*rl)
		} else {
			//fmt.Printf("Id doesn't need any params")
			fmt.Printf("-1 ")
		}
	case "exit":
		os.Exit(0)
	default:
		//fmt.Printf("NOT A VALID COMMAND")
	}
}

func main() {
	//shell grabs input but can be multiple groups max 5 words
	//splits each word into a bucket in array

	//id: none | in: 1-5 || cr: 1
	//de: 1 | rq: 1-2 | rl: 1-2 | to: none

	//list of pcbs, max size is 16 then has to reallocate
	var pcbArray []*pcb

	pcbArray = make([]*pcb, 16)

	for i := range pcbArray {
		pcbArray[i] = nil
	}

	//list of rcbs
	var rcbArray []*rcb

	rcbArray = make([]*rcb, 4)

	for i := range rcbArray {
		rcbArray[i] = nil
	}

	//initializing the ready list, will be fully defined by in or id
	var rl [][]int

	//============================================================================================

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.Split(line, " ")
		command := Command{Type: parts[0], Args: parts[1:]}
		//commands = append(commands, command)
		runCommand(&pcbArray, &rcbArray, &rl, command)
		//fmt.Printf("You entered:", line)
	}

}
