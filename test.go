package frostutil

import (
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func init() {
	runtime.LockOSThread()
}

// Any test function meant to run in Update must have this signature
type UpdateTestFunc func(t *testing.T)

// Any test function meant to run in Draw must have this signature
type DrawTestFunc func(t *testing.T, screen *ebiten.Image)

// Any test function meant to run in Layout must have this signature
type LayoutTestFunc func(t *testing.T, outsideWidth, outsideHeight int) (screenWidth, screenHeight int)

var updateTests chan *UpdateTest = make(chan *UpdateTest, 1)
var drawTests chan *DrawTest = make(chan *DrawTest, 1)
var layoutTests chan *LayoutTest = make(chan *LayoutTest, 1)
var awaitUpdateTestCompletion chan bool = make(chan bool)
var awaitDrawTestCompletion chan bool = make(chan bool)
var awaitLayoutTestCompletion chan bool = make(chan bool)
var hasTestMain bool // set to true by OnTestMain prior to calling m.Run(), if it is false in Queue*Test, then OnTestMain was never called.
var testsQueued int

// TestGame contains the Update, Layout, and Draw methods that Ebitengine calls.
type TestGame struct {
	screenWidth, screenHeight int
}

// This has to be called from a TestMain(m *testing.M) function in any package that uses QueueUpdateTest, QueueDrawTest, or QueueLayoutTest.
// It sets up and runs Ebitengine, runs your test functions (via m.Run) which should call Queue*Test, waits for it to finish,
// and then closes the channels and sets their variables to nil, which prompts Update to tell Ebitengine to shut down.
func OnTestMain(m *testing.M) {
	runtime.LockOSThread()
	f := func() {
		hasTestMain = true
		m.Run()
		close(updateTests)
		close(drawTests)
		close(layoutTests)
		drawTests = nil
		layoutTests = nil
		updateTests = nil
		for testsQueued > 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	go f()
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Test")
	//time.Sleep(time.Second)
	testGame := &TestGame{screenWidth: 1920, screenHeight: 1080}
	ebiten.RunGame(testGame)
	runtime.UnlockOSThread()
}

// UpdateTest pointers are sent through a channel from QueueUpdateTest to *TestGame.Update.
type UpdateTest struct {
	t *testing.T
	f UpdateTestFunc
}

// DrawTest pointers are sent through a channel from QueueDrawTest to *TestGame.Draw.
type DrawTest struct {
	t *testing.T
	f DrawTestFunc
}

// LayoutTest pointers are sent through a channel from QueueLayoutTest to *TestGame.Layout.
type LayoutTest struct {
	t *testing.T
	f LayoutTestFunc
}

// Each time Update is called by Ebitengine, it retrieves an update test, if any are queued, from the updateTests channel, runs it,
// and then lets QueueUpdateTest know that it has finished running it (so that it will return).
// If updateTests is nil, then it returns an error to tell Ebitengine to shut down.
func (game *TestGame) Update() (err error) {
	if updateTests != nil {
		if len(updateTests) > 0 {
			test := <-updateTests
			test.f(test.t)
			awaitUpdateTestCompletion <- true
		}
	} else {
		err = errors.New("Done")
	}
	return
}

// Each time Draw is called by Ebitengine, it retrieves a draw test, if any are queued, from the drawTests channel, runs it,
// and then lets QueueDrawTest know that it has finished running it (so that it will return).
// If drawTests is nil, it does nothing.
func (game *TestGame) Draw(screen *ebiten.Image) {
	if drawTests != nil {
		if len(drawTests) > 0 {
			test := <-drawTests
			test.f(test.t, screen)
			awaitDrawTestCompletion <- true
		}
	}
}

// Each time Layout is called by Ebitengine, it retrieves a layout test, if any are queued, from the layoutTests channel, runs it,
// records the screenWidth and screenHeight that it returns, and then lets QueueLayoutTest know that it has finished running it (so that it will return).
// If layoutTests is nil, it does nothing.
// It returns the screenWidth and screenHeight returned by the last layout test, or 1920 and 1080 if no layout tests were ever queued.
func (game *TestGame) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	if layoutTests != nil {
		if len(layoutTests) > 0 {
			test := <-layoutTests
			game.screenWidth, game.screenHeight = test.f(test.t, outsideWidth, outsideHeight)
			awaitLayoutTestCompletion <- true
		}
	}
	return game.screenWidth, game.screenHeight
}

// QueueUpdateTest checks to make sure OnTestMain was called, and if it was, it packages up the parameters t and f,
// and sends them through the updateTests channel for Update. It waits for Update to let it know that it has finished running f(t), and then returns.
// If OnTestMain was never called, it triggers a test failure and warns that you need to call OnTestMain from TestMain in every package which contains calls to QueueUpdateTest.
func QueueUpdateTest(t *testing.T, f func(t *testing.T)) {
	if hasTestMain {
		testsQueued++
		updateTests <- &UpdateTest{t, f}
		<-awaitUpdateTestCompletion
		testsQueued--
	} else {
		t.Fatal("Missing call to frostutil.OnTestMain. OnTestMain must be called from a TestMain(m *testing.M) function in every package in which you want to use QueueLayoutTest, QueueUpdateTest, and/or QueueDrawTest.")
	}
}

// QueueDrawTest checks to make sure OnTestMain was called, and if it was, it packages up the parameters t and f,
// and sends them through the drawTests channel for Draw. It waits for Draw to let it know that it has finished running f(t, screen), and then returns.
// If OnTestMain was never called, it triggers a test failure and warns that you need to call OnTestMain from TestMain in every package which contains calls to QueueDrawTest.
func QueueDrawTest(t *testing.T, f func(t *testing.T, screen *ebiten.Image)) {
	if hasTestMain {
		testsQueued++
		drawTests <- &DrawTest{t, f}
		<-awaitDrawTestCompletion
		testsQueued--
	} else {
		t.Fatal("Missing call to frostutil.OnTestMain. OnTestMain must be called from a TestMain(m *testing.M) function in every package in which you want to use QueueLayoutTest, QueueUpdateTest, and/or QueueDrawTest.")
	}
}

// QueueLayoutTest checks to make sure OnTestMain was called, and if it was, it packages up the parameters t and f,
// and sends them through the layoutTests channel for Layout. It waits for Layout to let it know that it has finished running f(t, outsideWidth, outsideHeight),
// and then returns.
// If OnTestMain was never called, it triggers a test failure and warns that you need to call OnTestMain from TestMain in every package which contains calls to QueueLayoutTest.
func QueueLayoutTest(t *testing.T, f func(t *testing.T, outsideWidth, outsideHeight int) (screenWidth, screenHeight int)) {
	if hasTestMain {
		testsQueued++
		layoutTests <- &LayoutTest{t, f}
		<-awaitLayoutTestCompletion
		testsQueued--
	} else {
		t.Fatal("Missing call to frostutil.OnTestMain. OnTestMain must be called from a TestMain(m *testing.M) function in every package in which you want to use QueueLayoutTest, QueueUpdateTest, and/or QueueDrawTest.")
	}
}
