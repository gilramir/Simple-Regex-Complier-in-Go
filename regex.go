package main

import(
	"fmt"
	"bufio"
	"os"
)

/*
 * Converts the regular expression to a postfix notation.
 * Pretty standard implementation of the postfix algorithm
 * aside from a few changes to include the concatenation operator
 */
func re2post(re string) ([]rune,int){
	buffer := make([]rune,8000)
	i:=0
	nbin:=0
	natom:=0
	type backc struct{
		nbin,natom int
	}
	p := make([]backc,1000)
	j:=0

	for _,ch := range re{
		switch ch{
		case '(':
			if natom>1{
				natom--
				buffer[i]='.'
				i++
			}
			p[j].nbin=nbin
			p[j].natom=natom
			j++;
			nbin=0
			natom=0

		case '|':
			if natom == 0{
				return nil,0
			}
			for natom--;natom>0;natom--{
				buffer[i]='.'
				i++
			}
			nbin++

		case ')':
			if j==0 || natom==0{
				return nil,0
			}
			for natom--;natom>0;natom-- {
				buffer[i]='.'
				i++
			}
			for ;nbin>0;nbin--{
				buffer[i]='|'
				i++
			}
			j--
			nbin=p[j].nbin
			natom=p[j].natom
			natom++

		case '*','+','?':
			if natom==0{
				return nil,0
			}
			buffer[i]=ch
			i++
		default:
			if natom>1{
				natom--
				buffer[i]='.'
				i++
			}
			buffer[i]=ch
			i++
			natom++
		}
	}
	if j!=0{
		return nil,0
	}
	for natom--;natom>0;natom--{
		buffer[i]='.'
		i++
	}
	for ; nbin>0;nbin--{
		buffer[i]='|'
		i++
	}

	return buffer,i
}


/*
 * Represents an NFA state plus zero or one or two arrows exiting.
 * if c == Match, no arrows out; matching state.
 * If c == Split, unlabeled arrows to out and out1 (if != NULL).
 */
const(
	Match = iota+1
	Split 
)
type State struct{
	c int
	s rune
	out, out1 *State
	lastlist int
}

/* Singleton matching state used to denote all end states*/
var matchstate = State{c:Match}
var nstate int
var listid int

/*
 * A partially built NFA without the matching state filled in.
 * Frag.start points at the start state.
 * Frag.out is a list of places that need to be set to the
 * next state for this fragment.
 */
type Frag struct{
	start *State
	out []**State
}

/* Patch the list of states at out to point to s. */
func patch(out []**State, s *State){
	for _,p := range out {
		*p = s
	}
}

/*
 * Convert postfix regular expression to NFA.
 * Return start state.
 */
func post2nfa(pf []rune,j int) *State{
	stack := make([]Frag,1000)
	stp :=0
	for i:=0;i<j;i++{
		switch pf[i]{
		default:
			s := State{s:pf[i],out:nil,out1:nil}
			stack[stp]=Frag{&s,[]**State{&s.out}}
			stp++
		case '.':
			stp--
			e2 := stack[stp]
			stp--
			e1 := stack[stp]
			/*catenate*/
			patch(e1.out, e2.start);
			stack[stp]=Frag{e1.start,e2.out}
			stp++
		case '|': /*alternate*/
			stp--
			e2 := stack[stp]
			stp--
			e1 := stack[stp]
			s := State{c:Split,out:e1.start,out1:e2.start}
			stack[stp]=Frag{&s,append(e1.out,e2.out...)}
			stp++
		case '?': /*zero or one*/
			stp--
			e := stack[stp]
			s := State{c:Split,out:e.start}
			stack[stp] =Frag{&s,append(e.out,&s.out1)}
			stp++
		case '*': /*zero or more*/
			stp--
			e := stack[stp]
			s := State{c:Split,out:e.start}
			patch(e.out, &s);
			stack[stp] =Frag{&s,[]**State{&s.out1}}
			stp++
		case '+': /*one or more*/
			stp--
			e := stack[stp]
			s := State{c:Split,out:e.start}
			patch(e.out, &s);
			stack[stp] =Frag{e.start,[]**State{&s.out1}}
			stp++
		}
	}
	stp--
	e := stack[stp]
	if stp != 0{
		return nil
	}
	patch(e.out,&matchstate)
	return e.start
}

/* Add s to l, following unlabeled arrows. */
func addstate(l []*State,s *State) []*State{
	if s==nil || s.lastlist == listid {
		return l
	}
	s.lastlist=listid
	if s.c==Split{
		// Add the 2 output States
		l=addstate(l,s.out)
		l=addstate(l,s.out1)
		// Return without adding the Split state again
		return l
	}
	l=append(l,s)
	return l
}

/*
 * Step the NFA from the states in clist
 * past the character ch,
 * to create next NFA state set nlist.
 */
func step(clist []*State,ch rune,nlist []*State) []*State{
	listid++;
	nlist = nlist[:0]
	for _,s := range clist{
		if s.s == ch{
			nlist=addstate(nlist,s.out)
		}
	}
	return nlist
}

/* Check whether state list contains a match. */
func ismatch(l []*State) bool{
	for _,s := range l{
		if s==&matchstate {
			return true
		}
	}
	return false
}


func match(start *State,s string ) bool{
	var clist,nlist []*State
	listid++;
	clist = addstate(clist,start)
	
	for _,ch := range s{
		nlist=step(clist,ch,nlist)
		clist,nlist = nlist,clist
	}
	return ismatch(clist);

}


func main(){
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Enter Regex: ")
	for scanner.Scan() {
		buf,j:=re2post(scanner.Text())
		if buf!=nil{
			//fmt.Println(string(buf))
		}else{
			fmt.Println("invalid regex")
			continue
		}
		start := post2nfa(buf,j)
		if(start==nil){
			fmt.Println("error while building nfa")
			continue
		}
		fmt.Printf("Enter String to match: ")
		scanner.Scan()
		if match(start,scanner.Text()){
			fmt.Println("match found!")
		}else{
			fmt.Println("no match found!")
		}
		fmt.Printf("Enter Regex: ")
	}
}