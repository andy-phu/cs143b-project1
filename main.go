package main

import (
	"fmt"
	"bufio"
	"strings"
	"os"
    "container/list"
	"strconv"
)

type pcb struct{
	State int
	Parent int
	Priority int //1 to n-1
	Children *list.List
	Resources *list.List
}

type rcb struct{
	State int
	Inventory int
	Waitlist *list.List
}

type waitlistProcess struct{
	Index int //index of the process in the pcb
	requestedUnits int
}

//pcb resources is a linked list of these 
type resourceInfo struct{
	ResourceIndex int 
	Units int
}

//finds the element inside of a linked list regardless of type of linked list 
func findElement(index int, list *list.List) interface{}{
	var counter int = 0;
	for e := list.Front() ; e != nil; e = e.Next(){
		if counter == index{
			return e.Value
		}
		counter++
	}
	return nil
}

//find the empty slot (index) in pcb to add a new process pcb
func findEmptyPCB(pcbArray []*pcb) int{
	for e := 0 ; e < len(pcbArray); e++{
		if(pcbArray[e] == nil){
			return e
		}
	
	}
	return -1 
}

//find the empty bucket in an array (mainly used for ready list)
func findEmptySlot(arr[] int)int{
	for i:=0; i < len(arr); i++{
		if (arr[i] != -1){
			return i
		}
	}
	return -1
}

//find the running process based on the head of the ready list, search top down from 2 (highest) 
//RETURN: index of the running process in the pcb array
func findRunningProcess(readyList *[][]int) int{
	n := len(*readyList)

	for i:=n; i > 0; i++{
		//the running process will always be the first element in the inner arrays
		if((*readyList)[i][0] != -1){
			return (*readyList)[i][0]
		}
	}
	//if all of them are empty, this is the beginning of init
	return -1
}

//finds the running process in ready list which iterates through each level and finds the level with a non null array
//removes the first element in that level
func readyListRemoval(readyList *[][]int){
	var n int = len(*readyList)

	//from n to 1 bc 0 doesnt have a running reserved for init
	for i:=n; i > 0; i--{
		//non null means that the running process is there
		if (*readyList)[i] != nil{
			//shifts over the head to remove 
			for j := 1; j < 16; j++ {
				(*readyList)[i][j-1] = (*readyList)[i][j]
			}
		}
	}
}

func resourceListRemoval(pcbArray *[]*pcb, pcbIndex int, resourceIndex int){
	for e:=(*pcbArray)[pcbIndex].Resources.Front(); e != nil; e = e.Next(){
		//find the node in the resources linked list and remove
		if e.Value.(int) == resourceIndex{
			(*pcbArray)[pcbIndex].Resources.Remove(e)
			eString := strconv.Itoa(e.Value.(int))
			fmt.Println("Removing resource index: " + eString)
			break
		}

	} 
}

func waitListRemoval(rcbArray *[]*rcb, resourceIndex int, requestedUnits int, pcbArray *[]*pcb, readyList *[][]int){
	for{
		// while (r.waitlist != empty and r.state > 0)
		if (((*rcbArray)[resourceIndex].Waitlist != nil) && ((*rcbArray)[resourceIndex].State > 0)){
			break;
		}

		// get next (j, k) from r.waitlist, j is the process index and k is the amt of resources requested by process
		var unblockedProcess waitlistProcess = (*rcbArray)[resourceIndex].Waitlist.Front().Value.(waitlistProcess)


		// if (r.state >= k)
		if (*rcbArray)[resourceIndex].State >= unblockedProcess.requestedUnits{
			// r.state = r.state - k
			(*rcbArray)[resourceIndex].State = (*rcbArray)[resourceIndex].State - unblockedProcess.requestedUnits
			// insert (r, k) into j.resources
			newResource := &resourceInfo{ResourceIndex: resourceIndex, Units: requestedUnits}
			(*pcbArray)[unblockedProcess.Index].Resources.PushBack(newResource)
			// j.state = ready
			(*pcbArray)[unblockedProcess.Index].State = 1
			// remove (j, k) from r.waitlist		
			var removed bool = false
			for e:=(*rcbArray)[resourceIndex].Waitlist.Front(); e != nil; e = e.Next(){
				node := e.Value.(waitlistProcess)
				if node == unblockedProcess{
					(*rcbArray)[resourceIndex].Waitlist.Remove(e)
					removed = true
				}
			}
			
			indexString := strconv.Itoa(unblockedProcess.Index)
			unitString := strconv.Itoa(unblockedProcess.requestedUnits)
			if removed{
				fmt.Println("Process: " + indexString + " Units: " + unitString + " SUCCESSFUL removal from WL")
			}else{
				fmt.Println("Process: " + indexString + " Units: " + unitString + " FAILED removal from WL")
			}

			//insert j into RL
			var newPrio int = (*pcbArray)[unblockedProcess.Index].Priority 
			(*readyList)[newPrio] = append((*readyList)[newPrio], unblockedProcess.Index)
		}else{
			break
		}
	}

}

func scheduler(readyList [][]int){
	var n int = len(readyList)

	//from n to 1 bc 0 doesnt have a running reserved for init
	for i:=n; i > 0; i--{
		//non null means that the running process is there
		if (readyList)[i] != nil{
			head := strconv.Itoa(readyList[i][0])
			fmt.Println("Process: " + head + " running")
			return 
		}
	}
	fmt.Println("SCHEDULER ERROR: No highest priority ready process found")
}

//Init: n = amt of priority levels | u_num = the amt of units for resource_num
//Notes: creates ready list with n priority levels 0 to n-1, and returns it
func in(n string, u0 string, u1 string, u2 string, u3 string, pcbArray *[]*pcb, rcbArray *[]*rcb)[][]int{
	var cmdLine string = fmt.Sprintf("p1: %s, p2: %s, p3: %s, p4: %s, p5: %s", n, u0, u1, u2, u3)
	fmt.Println("all of the inputs for in: " + cmdLine)

	var int0, _ = strconv.Atoi(u0)
	var int1, _ = strconv.Atoi(u1)
	var int2, _ = strconv.Atoi(u2)
	var int3, _ = strconv.Atoi(u3)


	prioLevels, _ := strconv.Atoi(n)

	if prioLevels<=0{
		fmt.Println("ERROR: must have at least 1 priority level")
		return nil
	}else{
		//initializes the rcb array with the params 
		var rcb0 rcb = rcb{
			State: int0,
			Inventory: int0,
			Waitlist: list.New(),
		}

		var rcb1 rcb = rcb{
			State: int1,
			Inventory: int1,
			Waitlist: list.New(),
		}

		var rcb2 rcb = rcb{
			State: int2,
			Inventory: int2,
			Waitlist: list.New(),
		}

		var rcb3 rcb = rcb{
			State: int3,
			Inventory: int3,
			Waitlist: list.New(),
		}
		
		lenString := strconv.Itoa(len(*rcbArray))
		fmt.Println("size of rcb array: " + lenString)
		(*rcbArray)[0] = &rcb0
		(*rcbArray)[1] = &rcb1
		(*rcbArray)[2] = &rcb2
		(*rcbArray)[3] = &rcb3

		//initializes a 2d ready list of n buckets with len 16 in each bucket
		readyList := make([][]int, prioLevels)

		for i:=0; i<prioLevels; i++{
			innerArray := make([]int, 16)
			for j := 0; j < 16; j++ {
				innerArray[j] = -1
			  }
			  readyList[i] = innerArray
		}
		
		//intializes the pcbArray
		create(&readyList, pcbArray, "0")
		fmt.Println("Successfully initialized!")
		scheduler(readyList)
		return readyList
	}
}

//Create: p = priority level (1,2,0 but 0 is for init process)
func create(readyList *[][]int, pcbArray *[]*pcb, p string){
	//allocate new PCB[j]
	//getes the empty slot to insert the new process pcb
	emptyPCB  := findEmptyPCB(*pcbArray)
	runningIndex := findRunningProcess(readyList)
	priority, _ := strconv.Atoi(p)

	//if there is no running process in ready list, this is init calling create
	if(runningIndex == -1){
		//creates the init pcb running with nothing at prio level 0 
		var newPCB pcb = pcb{
			State: -1,
			Parent: -1,
			Priority: 0,
			Children: list.New(),
			Resources:  list.New(),
		}

		//add to the pcbArray and to the readyList 
		(*pcbArray)[0] = &newPCB

		//ready list prio 0 at the head is index 0 of the init pcb 
		(*readyList)[0][0] =  0
	}else{
		//if there is a running process it is the one that calls create
		//assign the new pcb to the running process's child and vice versa new pcb parent = running 
		if(emptyPCB == -1){
			fmt.Println("ERROR: empty slot is -1, too many processes")
			return
		}else{
			var newPCB pcb = pcb{
				State: 1,
				Parent: runningIndex,
				Priority: priority,
				Children: list.New(),
				Resources: list.New(),
			}
	
			//updating the running process children list 
			(*pcbArray)[runningIndex].Children.PushBack(emptyPCB)
	
			//add the new process to pcb array
			(*pcbArray)[emptyPCB] = &newPCB
			
	
			if priority == 0{
				fmt.Println("ERROR: not init -> cannot add process in priority level 0")
				return 
			}
	
			emptySlot := findEmptySlot((*readyList)[priority])
	
			//add to readylist 
			(*readyList)[priority][emptySlot] = emptyPCB
	
			fmt.Println("Process: " + strconv.Itoa(emptyPCB) + " created successfully!")
		}
	}
	scheduler(*readyList)
}

//TODO
//Destroy: i = pcb index
func destroy(p string){
	
}

//Request: r = resource number | k = num of units for resource r
func request(readyList *[][]int, pcbArray *[]*pcb, rcbArray *[]*rcb, r string, k string)int{
	runningIndex := findRunningProcess(readyList)
	resourceNum, _ := strconv.Atoi(r)
	requestedUnits, _ := strconv.Atoi(k)
	inventory := (*rcbArray)[resourceNum].Inventory
	state := (*rcbArray)[resourceNum].State

	//num of units requested + num alr held <= initial inventory 
	//k + (inventory - state) <= inventory
	//fails check then automatically return -1 
	if (requestedUnits + (inventory - state) > inventory){
		fmt.Println("ERROR: requested more than the available units for resource: " + r)
		return -1
	}
	if (requestedUnits <= 0){
		fmt.Println("ERROR: the amount of units requested has to be greater than 0")
		return -1
	}

	// if state of r is free
	if(state > 0){
		// state of r = allocated
		(*rcbArray)[resourceNum].State = state - requestedUnits
		// insert r into list of resources of process i
		(*pcbArray)[runningIndex].Resources.PushBack(resourceNum)
		// display: “resource r allocated”
		fmt.Println("Resource: " + r + " allocated!")
		scheduler(*readyList)
		return 1 //1 is successfully allocated
	}else{
		// state of i = blocked -> state in pcb becomes 0 
		(*pcbArray)[runningIndex].State = 0
		//remove i (head of RL) from RL 
		readyListRemoval(readyList)
		//add (i,k) to waitlist of r 
		newResource := &resourceInfo{ResourceIndex: resourceNum, Units: requestedUnits}
		(*rcbArray)[resourceNum].Waitlist.PushBack(newResource)
		// display: “process i blocked”
		runningIndexString := strconv.Itoa(runningIndex)
		fmt.Println("Process: " + runningIndexString + " allocated!")
		scheduler(*readyList)	
		return 0 //0 is blocked
	}
	
	
}

//Release: r = resource number | k = num of units for resource r
func release(readyList *[][]int, pcbArray *[]*pcb, rcbArray *[]*rcb, r string, k string) int {
	runningIndex := findRunningProcess(readyList)
	resourceNum, _ := strconv.Atoi(r)
	requestedUnits, _ := strconv.Atoi(k)
	inventory := (*rcbArray)[resourceNum].Inventory
	state := (*rcbArray)[resourceNum].State

	//error check: num of units <= num of units currently held 
	//k <= (inventory - state)
	if (requestedUnits > (inventory-state)){
		fmt.Println("ERROR: released more than the number of units currently held for resource: " + r)
		return -1
	}
	
	// remove r from resources list of process i
	resourceListRemoval(pcbArray, runningIndex, resourceNum)

	// if waitlist of r is empty
	if (*rcbArray)[resourceNum].Waitlist == nil{
		// state of r = free (state goes back to original amt of units)
		(*rcbArray)[resourceNum].State = state + requestedUnits
	}else{
		//remove process j (head of WL) from the WL bc no longer blocked by process i -> joins RL
		waitListRemoval(rcbArray, resourceNum, requestedUnits, pcbArray, readyList)
	}
	scheduler(*readyList)
	return 1
}	

//Timeout
func timeout(readyList *[][]int){
	//store the og running process
	var runningProcess int = findRunningProcess(readyList)
	//iterate from n-1 to 1, and find the first array that isn't nil
	var runningPriority int;
	for i:=(len(*readyList)-1); i>0; i++{
		if (*readyList)[i] != nil{
			runningPriority = i
		}
	}
	//remove the running process and append to the back on the same prio level same array
	readyListRemoval(readyList)
	(*readyList)[runningPriority] = append((*readyList)[runningPriority], runningProcess)
	scheduler(*readyList)
}


func main() {
	//shell grabs input but can be multiple groups max 5 words 
	//splits each word into a bucket in array
	var cmd string
	
	//id: none | in: 1-5 || cr: 1 
	//de: 1 | rq: 1-2 | rl: 1-2 | to: none 
	var p1, p2, p3, p4, p5 string

	//list of pcbs, max size is 16 then has to reallocate
	var pcbArray []*pcb 

	pcbArray = make([]*pcb, 0, 16)

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

	for {
		reader := bufio.NewReader(os.Stdin) 
	
		fmt.Print("> ")
		
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line) 
	
		input := strings.Split(line, " ")
	
		// fmt.Println("Output: " + line)
		// fmt.Println("Input arr:", input)
		cmd = input[0]
		
		switch cmd{
		
		case "in":
			fmt.Println("Command: " + cmd)
			if(len(input) == 6){
				p1 = input[1]
				p2 = input[2]
				p3 = input[3]
				p4 = input[4]
				p5 = input[5]

				
				rl = in(p1,p2,p3,p4,p5, &pcbArray, &rcbArray)
			}else{
				fmt.Println("Not enough params for the cmd: in")
			}
		case "cr":
			fmt.Println("Command: " + cmd)
			if(len(input) == 2){
				p1 = input[1]

				create(&rl, &pcbArray, p1 )
			}else{
				fmt.Println("Not enough params for the cmd: cr")
			}
		case "de":
			fmt.Println("Command: " + cmd)
			if(len(input) == 2){
				p1 = input[1]
				
				destroy(p1)
			}else{
				fmt.Println("Not enough params for the cmd: de")
			}
		case "rq":
			fmt.Println("Command: " + cmd)

			if(len(input) == 3){
				p1 = input[1]	
				p2 = input[2]
				
				if request(&rl, &pcbArray, &rcbArray, p1, p2) == -1{
					fmt.Println("ERROR: failed to request")
				}
			}else{
				fmt.Println("Not enough params for the cmd: rq")
			}

		case "rl":
			fmt.Println("Command: " + cmd)

			if(len(input) == 3){
				p1 = input[1]	
				p2 = input[2]
				
				if release(&rl, &pcbArray, &rcbArray, p1, p2) == -1{
					fmt.Println("ERROR: failed to release")
				}
			}else{
				fmt.Println("Not enough params for the cmd: rl")
			}
		case "to":
			fmt.Println("Command: " + cmd)

			if(len(input) == 1){
				timeout(&rl)
			}else{
				fmt.Println("Not enough params for the cmd: rq")
			}
		case "id":
			fmt.Println("Command: " + cmd)
			if(len(input) == 1){
				in("3", "1", "1", "2", "3", &pcbArray, &rcbArray)
			}else{
				fmt.Println("Id doesn't need any params")
			}
		case "exit":
			os.Exit(0)
		default:
			fmt.Println("NOT A VALID COMMAND")
		}
		
		var runningProcess int = findRunningProcess(&rl)
		runningString := strconv.Itoa(runningProcess)

		fmt.Println("=================================")
		fmt.Println("Running Process: " + runningString)
		fmt.Println("=================================")


	}
}
