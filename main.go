// B1NukeBomber project main.go
package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/abiosoft/ishell/v2"
	"moul.io/banner"

	"container/list"

	"github.com/fatih/color"
	"github.com/gocarina/gocsv"
	"github.com/mmcloughlin/spherand"
	geo "github.com/rbsns/golang-geo"
	"github.com/scylladb/termtables"
)

const VERSION string = "0.4"

//Thule Air Base
//{"DD":{"lat":76.52533,"lng":-68.702},"DMS":{"lat":"76º31'31.19\" N","lng":"68º42'7.19\" W"},"geohash":"fmx5keh8r7","UTM":"19X 507751.39524828 8493824.41212883"}

/*
Speed:	Max Speed: Mach 1.2 at sea level (900+ mph)
Max Speed: Mach 2.3 at 50,000 feet (1,450 mph / 1,259 knots)
Cruise speed: 560 mph (487 knots) to 650 mph (1040 Km/H / 560 Kt) 0.28 Km/sec
*/
const CRUISE_SPEED float32 = 1040.0
const MAX_FUEL = 120326.0            //https://www.boeing.com/defense/b-1b-bomber/
const ProbabilityOfKillSam200 = 0.85 //https://en.wikipedia.org/wiki/S-200_(missile)
const MAX_ALTITUDE = 18000.0         //Cealing More than 30,000 ft (9,144 m)
const MIN_ALTITUDE = 80.0            //80 m
const FUEL_CONSUMPTION = 2.51        //Kg/sec

/*
https://en.wikipedia.org/wiki/Probability_of_kill
The probability of kill, or "Pk", is usually based on a uniform random number generator.
This algorithm creates a number between 0 and 1 that is approximately uniformly distributed in that space.
If the Pk of a weapon/target engagement is 30% (or 0.30), then every random number generated that is less than 0.3
is considered a "kill"; every number greater than 0.3 is considered a "no kill".
When used many times in a simulation, the average result will be that 30% of the weapon/target engagements
will be a kill and 70% will not be a kill.
*/

//General var
var T0 time.Time
var Tnew time.Time
var DeltaT time.Duration

//Player B1 Bomber
type bomber struct {
	lat      float64
	long     float64
	fuel     int
	speed    float32
	bearing  float32
	altitude int
	ecm      bool
	target   string
	targetAb string
	safecode string
}

//Targets
type Target struct {
	name         string
	abbreviation string
	lat          float64
	long         float64
	targetype    string
}

type TargetCSV struct { // Our example struct, you can use "-" to ignore a field
	Name         string `csv:"target_name"`
	Abbreviation string `csv:"target_ab"`
	Lat          string `csv:"target_lat"`
	Long         string `csv:"target_long"`
	Type         string `csv:"target_type"`
	NotUsed      string `csv:"-"`
}

var b1 bomber

//https://stackoverflow.com/questions/53931002/using-a-global-list-variable-in-golang-receiving-use-of-package-list-without-s/53931180
var ListTargets *list.List

var letters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

//https://stackoverflow.com/a/22892986
func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func StringToInt(data string) int {
	n, _ := strconv.Atoi(data)
	return n
}

func IntToString(n int) string {
	return strconv.Itoa(n)
}

func StringToFloat32(data string) float32 {
	if s, err := strconv.ParseFloat(data, 32); err == nil {
		return float32(s)
	} else {
		return 0.0
	}
}

func StringToFloat64(data string) float64 {
	if s, err := strconv.ParseFloat(data, 64); err == nil {
		return s
	} else {
		return 0.0
	}
}

//https://stackoverflow.com/a/37247762
func round(val float64) int {
	if val < 0 {
		return int(val - 0.5)
	}
	return int(val + 0.5)
}

//https://go.dev/play/p/Q-Ufgrw3vZL
func DDHHMMZ() string {
	current_time := time.Now().UTC()
	return fmt.Sprintf("%d%02d%02dZ\n", current_time.Day(), current_time.Hour(), current_time.Minute())
}

func DDHHMMZmmmYY() string {
	current_time := time.Now().UTC()
	return fmt.Sprintf(current_time.Format("021504ZJan06\n"))
}

//https://stackoverflow.com/questions/49746992/generate-random-float64-numbers-in-specific-range-using-golang
func randFloats(min, max float32, n int) []float32 {
	res := make([]float32, n)
	for i := range res {
		res[i] = min + rand.Float32()*(max-min)
	}
	return res
}

// `newBomber` constructs a new person struct with the given name.
func newBomber(fuel int) *bomber {
	// You can safely return a pointer to local variable
	// as a local variable will survive the scope of the function.
	b := bomber{fuel: fuel}
	b.speed = CRUISE_SPEED
	return &b
}

// Main function to move Player B1 Bomber
func Moving() {
	Tnew = time.Now().UTC()
	DeltaT = Tnew.Sub(T0)
	T0 = Tnew
	DistanceNew := float32(DeltaT.Seconds()) * (b1.speed / 3600) //Km/sec
	b1.fuel = b1.fuel - round(DeltaT.Seconds()*FUEL_CONSUMPTION)
	p_b1 := geo.NewPoint(b1.lat, b1.long)
	p_b1_new := p_b1.PointAtDistanceAndBearing(float64(DistanceNew), float64(b1.bearing))
	b1.lat = p_b1_new.Lat()
	b1.long = p_b1_new.Lng()
}

// Show present status B1 Bomber
func CheckB1Status() string {
	Moving()
	s := fmt.Sprintf("Present status:\n")
	s = s + fmt.Sprintf("Lat: %f\n", b1.lat)
	s = s + fmt.Sprintf("Long: %f\n", b1.long)
	s = s + fmt.Sprintf("Course: %f\n T", b1.bearing)
	s = s + fmt.Sprintf("Fuel: %d\n Kg", b1.fuel)
	s = s + fmt.Sprintf("Speed: %.2f Km/H\n", b1.speed)
	s = s + fmt.Sprintf("Altitude: %d m\n", b1.altitude)
	s = s + fmt.Sprintf("DeltaT time game: %s\n", DeltaT.String()) //DEBUG
	if b1.ecm {
		s = s + fmt.Sprintf("ECM: ON\n")
	} else {
		s = s + fmt.Sprintf("ECM: OFF\n")
	}
	s = s + fmt.Sprintf("YOUR PRIMARY TARGET IS: (%s) %s\n", b1.targetAb, b1.target)
	return s
}

//Show all navigations info
//See: https://dev.to/shumito/using-lists-in-go-with-structs-5gbk
func Navigation(data string) string {
	Moving()
	table := termtables.CreateTable()
	table.AddHeaders("Target Name", "Target Abbr.", "Distance", "Course")
	p_b1 := geo.NewPoint(b1.lat, b1.long)
	for e := ListTargets.Front(); e != nil; e = e.Next() {
		itemTarget := Target(e.Value.(Target))
		if itemTarget.targetype == "C" {
			p_target := geo.NewPoint(itemTarget.lat, itemTarget.long)
			dist := p_b1.GreatCircleDistance(p_target)
			bearing := p_b1.BearingTo(p_target)
			//Fix: https://github.com/rbsns/golang-geo/commit/8004ce49479db5787a6cab51b37b38e9ce052ca5
			if bearing < 0. {
				bearing = 360. + bearing
			}
			if len(data) == 0 {
				//s = s + fmt.Sprintf("Name: "+itemTarget.name+" Abbr.: "+itemTarget.abbreviation+" Dist.: %.2f Km\n", dist)
				table.AddRow(itemTarget.name, itemTarget.abbreviation, fmt.Sprintf("%.2f Km", dist), fmt.Sprintf("%.2f", bearing))
			} else {
				if itemTarget.name == data || itemTarget.abbreviation == data {
					table.AddRow(itemTarget.name, itemTarget.abbreviation, fmt.Sprintf("%.2f Km", dist), fmt.Sprintf("%.2f", bearing))
				}
			}
		}
	}
	return table.Render()
}

func main() {
	// create new shell.
	// by default, new shell includes 'exit', 'help' and 'clear' commands.
	shell := ishell.New()

	//Colors
	//cyan := color.New(color.FgCyan).SprintFunc()
	//yellow := color.New(color.FgYellow).SprintFunc()
	boldRed := color.New(color.FgRed, color.Bold).SprintFunc()

	// display welcome info.
	shell.Println(banner.Inline("nuke bomber"))
	shell.Println()
	shell.Println("** B-1 NUKE BOMBER GAME **")
	shell.Println(("By Enrico Speranza (Vytek). Version: " + VERSION))
	shell.Println()

	// display game info
	shell.Println("YOU ARE FLYING A B1 BOMBER OUT OF THULE AFB. YOU ARE IN AN ALERT STATUS")
	shell.Println("ORBITING OVER THE ARCTIC.")
	shell.Printf(boldRed("***** FLASH *****  HOT WAR HOT WAR HOT WAR"))
	shell.Println()
	shell.Println()

	T0 = time.Now().UTC()

	//Calculate diff: https://golangbyexample.com/time-difference-between-two-time-value-golang/
	shell.Println("T0 start time game: " + T0.String())

	/*
		Tnew := time.Now().UTC()
		DeltaT = Tnew.Sub(T0)
		T0 = Tnew
		shell.Println("DeltaT time game: " + DeltaT.String())
	*/

	//Random Lat Long for B1 Bomber
	g := spherand.NewGenerator(rand.New(rand.NewSource(1)))
	//fmt.Println(g.Geographical())
	lat0, long0 := g.Geographical()

	rand.Seed(time.Now().UnixNano())
	shell.Println(randFloats(1.10, 101.98, 5))
	bearing0 := randFloats(1.10, 101.98, 5)

	//Instance B1 Bomber
	b1 = bomber{lat: lat0, long: long0, fuel: MAX_FUEL, speed: CRUISE_SPEED, bearing: bearing0[0], altitude: int(25000 * rand.Float64())}

	//https://go.dev/play/p/n7jt1x3iw4Z
	//fmt.Println(b1) //DEBUG instance b1

	//CheckB1Status() //DEBUG

	//Load targets
	ListTargets = list.New()

	//LoadCSV
	csvFile, err := os.OpenFile("data/targets_db.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	targets := []*TargetCSV{}

	if err := gocsv.UnmarshalFile(csvFile, &targets); err != nil { // Load targets from file
		panic(err)
	}

	//Load targets from file and add to list //DEBUG
	for _, target := range targets {
		s_lat, _ := strconv.ParseFloat(target.Lat, 64)
		s_long, _ := strconv.ParseFloat(target.Long, 64)
		ListTargets.PushBack(Target{name: target.Name, abbreviation: target.Abbreviation, lat: s_lat, long: s_long, targetype: target.Type})
		shell.Println("Target", target.Name)
	}

	//Show all list in memory //DEBUG
	// Iterate the list
	CityType := 0
	DefenceType := 0
	for e := ListTargets.Front(); e != nil; e = e.Next() {
		itemTarget := Target(e.Value.(Target))
		//DEBUG
		shell.Println("Name: " + itemTarget.name + " Abbr.: " + itemTarget.abbreviation + " Lat: " + strconv.FormatFloat(itemTarget.lat, 'f', 5, 64) + " Long: " + strconv.FormatFloat(itemTarget.long, 'f', 5, 64))
		if itemTarget.targetype == "C" {
			CityType = CityType + 1
		}
		if itemTarget.targetype == "D" {
			DefenceType = DefenceType + 1
		}
	}

	//Choose a random target in ListTarget
	//https://golang.cafe/blog/golang-random-number-generator.html
	//https://pkg.go.dev/container/list#List.Len
	rand.Seed(time.Now().UnixNano())

	//Primary target.
	SelectTarget := rand.Intn(CityType + 1)
	shell.Printf("Number selected: %d\n", SelectTarget)
	i := 0
	for e := ListTargets.Front(); e != nil; e = e.Next() {
		itemTarget := Target(e.Value.(Target))
		if i == SelectTarget {
			b1.target = itemTarget.name
			b1.targetAb = itemTarget.abbreviation
		}
		i = i + 1
	}

	shell.Printf("YOUR PRIMARY TARGET IS: %s\n", b1.target)
	b1.safecode = randSeq(5)
	shell.Printf("YOUR FAIL SAFE CODE IS: %s\n", b1.safecode)

	// register a function for status" command.
	shell.AddCmd(&ishell.Cmd{
		Name: "status",
		Help: "Show status info about B1 Bomber",
		Func: func(c *ishell.Context) {
			//c.Println("Hello", strings.Join(c.Args, " "))
			c.Println(CheckB1Status())
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "st",
		Help: "(Abbreviation) Show status info about B1 Bomber",
		Func: func(c *ishell.Context) {
			c.Println(CheckB1Status())
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "datetime",
		Help: "Zulu datetime",
		Func: func(c *ishell.Context) {
			c.Println("Zulu time:", DDHHMMZ())
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "ldatetime",
		Help: "Zulu long datetime",
		Func: func(c *ishell.Context) {
			c.Println("Zulu time:", DDHHMMZmmmYY())
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "navigation",
		Help: "Show navigation info",
		Func: func(c *ishell.Context) {
			if len(c.Args) > 0 {
				c.Println(Navigation(strings.ToUpper(c.Args[0])))
			} else {
				c.Println(Navigation(""))
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "na",
		Help: "(Abbreviation) Show navigation info",
		Func: func(c *ishell.Context) {
			if len(c.Args) > 0 {
				c.Println(Navigation(strings.ToUpper(c.Args[0])))
			} else {
				c.Println(Navigation(""))
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "al",
		Help: "(Abbreviation) Change altitude",
		Func: func(c *ishell.Context) {
			if len(c.Args) > 0 {
				if StringToInt(c.Args[0]) > MAX_ALTITUDE {
					c.Printf("MAX Ceiling is: %s\n m", IntToString(MAX_ALTITUDE))
				} else if StringToInt(c.Args[0]) < MIN_ALTITUDE {
					c.Println("Min altitude is %s\n m", IntToString((MIN_ALTITUDE)))
				} else {
					b1.altitude = StringToInt(c.Args[0])
					c.Println(CheckB1Status())
				}
			} else {
				c.Println("Please, insert your new altitude")
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "altitude",
		Help: "Change altitude",
		Func: func(c *ishell.Context) {
			if len(c.Args) > 0 {
				if StringToInt(c.Args[0]) > MAX_ALTITUDE {
					c.Printf("MAX Ceiling is: %s\n m", IntToString(MAX_ALTITUDE))
				} else if StringToInt(c.Args[0]) < MIN_ALTITUDE {
					c.Println("Min altitude is %s\n m", IntToString((MIN_ALTITUDE)))
				} else {
					b1.altitude = StringToInt(c.Args[0])
					c.Println(CheckB1Status())
				}
			} else {
				c.Println("Please, insert your new altitude")
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "co",
		Help: "(Abbreviation) Change course",
		Func: func(c *ishell.Context) {
			if len(c.Args) > 0 {
				b1.bearing = StringToFloat32(c.Args[0])
				c.Println(CheckB1Status())
			} else {
				c.Println("Please, insert your new course")
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "course",
		Help: "Change course",
		Func: func(c *ishell.Context) {
			if len(c.Args) > 0 {
				b1.bearing = StringToFloat32(c.Args[0])
				c.Println(CheckB1Status())
			} else {
				c.Println("Please, insert your new course")
			}
		},
	})

	// run shell
	shell.Run()
}
