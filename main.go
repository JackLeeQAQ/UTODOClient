package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/objc"
)

func main() {
	runtime.LockOSThread()

	cocoa.TerminateAfterWindowsClose = false
	app := cocoa.NSApp_WithDidLaunch(func(n objc.Object) {
		obj := cocoa.NSStatusBar_System().StatusItemWithLength(cocoa.NSVariableStatusItemLength)
		obj.Retain()
		obj.Button().SetTitle("UTODO")

		nextClicked := make(chan bool)
		go func() {
			state := -1
			timer := 1500
			countdown := false
			for {
				select {
				case <-time.After(1 * time.Second):
					if timer > 0 && countdown {
						timer = timer - 1
					}
					if timer <= 0 && state%2 == 1 {
						state = (state + 1) % 4
						//当state=1且计时器结束时，发送HTTP请求
						if state == 2 {
							go func() {
								record := map[string]int64{
									"Id":   1,
									"Time": int64(10), // 获取WorkTime
								}
								body, _ := json.Marshal(record)
								resp, err := http.Post("http://localhost:36044?Action=CreateStudyTime", "application/json", bytes.NewBuffer(body))
								if err != nil {
									fmt.Println("Error upload the study record:", err)
									return
								}
								defer resp.Body.Close()
								fmt.Println("Study record:", resp.Status)
							}()
						}
					}
				case <-nextClicked:
					state = (state + 1) % 4
					timer = map[int]int{
						0: 2700,
						1: 10,
						2: 0,
						3: 600,
					}[state]
					if state%2 == 1 {
						countdown = true
					} else {
						countdown = false
					}
				}
				labels := map[int]string{
					0: "▶️ Ready %02d:%02d",
					1: "✴️ Working %02d:%02d",
					2: "✅ Finished %02d:%02d",
					3: "⏸️ Break %02d:%02d",
				}
				// updates to the ui should happen on the main thread to avoid strange bugs
				core.Dispatch(func() {
					obj.Button().SetTitle(fmt.Sprintf(labels[state], timer/60, timer%60))

				})
			}
		}()
		nextClicked <- true

		itemNext := cocoa.NSMenuItem_New()
		itemNext.SetTitle("Next")
		itemNext.SetAction(objc.Sel("nextClicked:"))
		cocoa.DefaultDelegateClass.AddMethod("nextClicked:", func(_ objc.Object) {
			nextClicked <- true
		})

		itemQuit := cocoa.NSMenuItem_New()
		itemQuit.SetTitle("Quit")
		itemQuit.SetAction(objc.Sel("terminate:"))
		menu := cocoa.NSMenu_New()
		menu.AddItem(itemNext)
		menu.AddItem(itemQuit)
		obj.SetMenu(menu)

	})
	app.Run()
}
