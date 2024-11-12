package main

import (
	"image"
	"image/color"
	w "lodestar/widgets"

	"github.com/AllenDang/cimgui-go/imgui"
	g "github.com/AllenDang/giu"
)

type Item struct {
	Id   string
	Long string
}

type Room struct {
	Path        string
	Includes    []string
	Inherit     string
	Name        string
	Short       string
	Long        string
	Smell       string
	Listen      string
	Items       []Item
	Exits       []string
	HiddenExits []string
}

type NodeCreationWindowData struct {
	Id    string
	Color color.RGBA
}

type EdgeCreationWindowData struct {
	Id            string
	DirectionName string
	From          string
	To            string
}

// Globals
// TODO: Use channels to share these across goroutines or something. This is bad.
var roomGraph *w.DraggableGraphWidget
var nodeCreationDialogues []*NodeCreationWindowData // TODO: Polymorphism so that we don't need 80000 arrays
var edgeCreationDialogues []*EdgeCreationWindowData

func nodeCreationMenu(graph *w.DraggableGraphWidget) {
	newNodeCreationDialogue := NodeCreationWindowData{
		Color: color.RGBA{R: 255, G: 255, B: 255, A: 255},
	}

	nodeCreationDialogues = append(nodeCreationDialogues, &newNodeCreationDialogue)
}

func renderNodeCreationMenu(window *g.WindowWidget, windowData *NodeCreationWindowData, windowDataIndex int, graph *w.DraggableGraphWidget, graphCenter image.Point, update *bool) {
	window.Layout(
		g.Column(
			g.Row(
				g.Label("Room ID"),
				g.InputText(&windowData.Id),
			),
			g.Row(
				g.Label("Room Color"),
				g.ColorEdit("", &windowData.Color),
			),
			g.Row(
				g.Button("Confirm").OnClick(func() {
					roomGraph.CreateNode(windowData.Id, graphCenter, windowData.Color, clicked, doubleClicked)
					nodeCreationDialogues = append(nodeCreationDialogues[:windowDataIndex], nodeCreationDialogues[windowDataIndex+1:]...)
					*update = true
				}),
				g.Button("Cancel").OnClick(func() {
					nodeCreationDialogues = append(nodeCreationDialogues[:windowDataIndex], nodeCreationDialogues[windowDataIndex+1:]...)
					*update = true
				}),
			),
		),
	)
}

func edgeCreationMenu(graph *w.DraggableGraphWidget, from string, to string) {
	newEdgeCreationDialogue := EdgeCreationWindowData{
		Id:            "",
		DirectionName: "",
		From:          from,
		To:            to,
	}

	edgeCreationDialogues = append(edgeCreationDialogues, &newEdgeCreationDialogue)

	g.Update()
}

func renderEdgeCreationMenu(window *g.WindowWidget, windowData *EdgeCreationWindowData, windowDataIndex int, graph *w.DraggableGraphWidget, update *bool) {
	window.Layout(
		g.Column(
			g.Label("From: "+windowData.From),
			g.Label("To: "+windowData.To),
			g.Row(
				g.Label("Exit Name:"),
				g.InputText(&windowData.DirectionName),
			),
			g.Row(
				g.Button("Confirm").OnClick(func() {
					roomGraph.CreateEdge(windowData.From+" "+windowData.DirectionName+" to "+windowData.To, windowData.From, windowData.To)
					edgeCreationDialogues = append(edgeCreationDialogues[:windowDataIndex], edgeCreationDialogues[windowDataIndex+1:]...)
					*update = true
				}),
				g.Button("Cancel").OnClick(func() {
					edgeCreationDialogues = append(edgeCreationDialogues[:windowDataIndex], edgeCreationDialogues[windowDataIndex+1:]...)
					*update = true
				}),
			),
		),
	)
}

func doubleClicked() {
}

func clicked() {
}

func dragged(offset image.Point) {
	return
}

func mainLoop() {
	// Just-in-case
	manualUpdateNeeded := false

	// Set up transparency
	imgui.PushStyleVarFloat(imgui.StyleVarWindowBorderSize, 0)
	g.PushColorWindowBg(color.RGBA{0, 0, 0, 150})
	g.PushColorFrameBg(color.RGBA{0, 0, 0, 0})

	// Get viewport size
	vX, vY := imgui.MainViewport().Size().X, imgui.MainViewport().Size().Y
	oX, oY := imgui.MainViewport().Pos().X, imgui.MainViewport().Pos().Y

	// Actually render our windows
	/// Render sub-windows
	//// Edge creation dialogue rendering
	for i, windowData := range edgeCreationDialogues {
		window := g.Window(windowData.From + " -> " + windowData.To)
		window.Size(vX/5, vY/3)
		window.Pos(oX+vX/2-vX/10, oY+vY/2-vY/6)
		renderEdgeCreationMenu(window, windowData, i, roomGraph, &manualUpdateNeeded)
	}

	//// Node creation dialogue rendering
	for i, windowData := range nodeCreationDialogues {
		window := g.Window("Node Creation")
		window.Size(vX/4, vY/3)
		window.Pos(oX+vX/2-vX/8, oY+vY/2-vY/6)
		graphCenter := roomGraph.GetOffset().Mul(-1).Add(image.Point{X: int(vX), Y: int(vY)}.Div(2))
		renderNodeCreationMenu(window, windowData, i, roomGraph, graphCenter, &manualUpdateNeeded)
	}

	/// Render main window
	mainWindow := g.Window("Lodestar")
	mainWindow.Flags(g.WindowFlagsNoDecoration | g.WindowFlagsNoMove | g.WindowFlagsNoBringToFrontOnFocus)
	mainWindow.Size(vX, vY)
	mainWindow.Layout(roomGraph)

	// Pop our styles so we don't crash
	g.PopStyleColor()
	g.PopStyleColor()
	g.PopStyle()

	// Just in case
	if manualUpdateNeeded {
		g.Update()
	}
}

func main() {
	mainPane := g.NewMasterWindow("Map", 800, 600, g.MasterWindowFlagsFloating|g.MasterWindowFlagsFrameless|g.MasterWindowFlagsTransparent)
	mainPane.SetBgColor(color.Transparent)

	// State-retained widgets
	roomGraph = w.DraggableGraph("room_graph", dragged, nodeCreationMenu, edgeCreationMenu)

	// GO! GO! GO! //
	mainPane.Run(mainLoop)
}
