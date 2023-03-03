package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
	"go.i3wm.org/i3/v4"
)

var debug bool

func init() {
	flag.BoolVar(&debug, "debug", false, "Enable debug log printing")
}

func waitForDbus() error {
	result := make(chan error, 1)
	go func() {
		result <- checkDbus()
	}()
	select {
	case <-time.After(10 * time.Second):
		return errors.New("timed out")
	case result := <-result:
		return result
	}
}

func checkDbus() error {
	conn, err := dbus.SessionBus()
	if err != nil {
		return err
	}
	defer conn.Close()

	for {
		obj := conn.Object("i3.status.rs", "/LayoutMode")
		call := obj.Call("i3.status.rs.SetStatus", 0, "", "", "Idle")
		if call.Err != nil {
			if debug {
				log.Println("Call response:", call.Err)
			}
		} else {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func main() {
	flag.Parse()

	err := waitForDbus()
	if err != nil {
		log.Fatalln("Error waiting for dbus:", err)
	}

	var wg sync.WaitGroup
	var lastFocus *i3.Node
	var previousFocus *i3.Node

	messages := make(chan string)
	wg.Add(2)
	go func() {
		defer wg.Done()
		recv := i3.Subscribe(i3.WindowEventType, i3.BindingEventType, i3.WorkspaceEventType)
		defer recv.Close()
		defer func() {
			messages <- "q"
		}()

		for recv.Next() {
			ev := recv.Event()
			switch ev := ev.(type) {
			case *i3.WorkspaceEvent:
				// log.Println("WorkspaceEvent:", ev.Change)
				if ev.Change == "init" {
					if debug {
						log.Println("New workspace using splith")
					}
					messages <- "splith"
				}
			case *i3.WindowEvent:
				// log.Println("WindowEvent:", ev.Change)
				if ev.Change == "focus" {
					if lastFocus != nil {
						previousFocus = lastFocus
					}
					focusedNode := ev.Container.FindChild(func(n *i3.Node) bool {
						return n.Focused
					})
					if focusedNode != nil {
						if parent := focusedNode.FindParent(); parent != nil {
							if debug {
								log.Println("focus change:", parent.Layout)
							}
							messages <- string(focusedNode.FindParent().Layout)
						} else {
							if debug {
								log.Println("No parent", focusedNode.Name)
							}
						}
						lastFocus = focusedNode
					} else {
						if debug {
							log.Println("No focused")
						}
					}
				} else {
					if debug {
						log.Println("change:", ev.Change)
					}
				}
			case *i3.BindingEvent:
				// log.Println("BindingEvent:", ev.Change)
				if ev.Change == "run" {
					switch ev.Binding.Command {
					case "layout tabbed":
						if debug {
							log.Println("Binding:", ev.Binding.Command)
						}
						messages <- "tabbed"
					case "layout stacking":
						if debug {
							log.Println("Binding:", ev.Binding.Command)
						}
						messages <- "stacked"
					case "layout toggle split":
						if debug {
							log.Println("Binding:", ev.Binding.Command)
						}
						messages <- "split"
					case "split v":
						if debug {
							log.Println("Binding:", ev.Binding.Command)
						}
						messages <- "splitv"
					case "split h":
						if debug {
							log.Println("Binding:", ev.Binding.Command)
						}
						messages <- "splith"
					case "nop last_focus":
						if debug {
							if previousFocus != nil {
								log.Println("previous_focus:", previousFocus.ID)
							}
							if lastFocus != nil {
								log.Println("last_focus:", lastFocus.ID)
							}
						}
						if previousFocus != nil {
							command := fmt.Sprintf("[con_id=\"%d\"] focus", previousFocus.ID)
							if debug {
								log.Println("Command:", command)
							}
							_, err := i3.RunCommand(command)
							if err != nil {
								log.Println("RunCommand error:", err)
							}
						}
					}
				}
			default:
				if debug {
					log.Println("Unknown message:", ev)
				}
			}

		}
	}()
	go func() {
		defer wg.Done()
		conn, err := dbus.SessionBus()
		if err != nil {
			log.Fatalln("Failed to connect to session bus:", err)
		}
		defer conn.Close()

		obj := conn.Object("i3.status.rs", "/LayoutMode")
		for msg := range messages {
			if msg == "q" {
				return
			}
			call := obj.Call("i3.status.rs.SetStatus", 0, msg, "", "Idle")
			if call.Err != nil {
				log.Fatalln("Failed to call function:", call.Err)
			}
			if debug {
				log.Println("dbus send:", msg)
			}
		}
	}()
	wg.Wait()
}
