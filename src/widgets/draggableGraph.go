package widgets

import (
	"image"
	"image/color"
	"slices"

	g "github.com/AllenDang/giu"
)

func isPointInRect(p image.Point, min image.Point, max image.Point) bool {
	if min.X <= p.X && p.X <= max.X && min.Y <= p.Y && p.Y <= max.Y {
		return true
	}

	return false
}

func findNodeUnderCursor(mouse image.Point, nodes []*GraphNodeWidget, nodeSize image.Point, offset image.Point) *GraphNodeWidget {
	for _, node := range nodes {
		if isPointInRect(mouse, node.position.Add(offset), node.position.Add(offset).Add(nodeSize)) {
			return node
		}
	}

	return nil
}

type GraphNodeWidget struct {
	id              string
	position        image.Point
	associatedEdges map[string]*GraphEdge
	color           color.Color
	clicked         func()
	doubleClicked   func()
}

func GraphNode(id string, position image.Point, color color.Color, clicked func(), doubleClicked func()) *GraphNodeWidget {
	return &GraphNodeWidget{
		id:              id,
		position:        position,
		associatedEdges: map[string]*GraphEdge{},
		color:           color,
		clicked:         clicked,
		doubleClicked:   doubleClicked,
	}
}

func (n *GraphNodeWidget) HasEdgeWith(target *GraphNodeWidget) *GraphEdge {
	for _, edge := range n.associatedEdges {
		if edge.from == target || edge.to == target {
			return edge
		}
	}

	return nil
}

func (n *GraphNodeWidget) HasEdgeTo(target *GraphNodeWidget) *GraphEdge {
	for _, edge := range n.associatedEdges {
		if edge.to == target {
			return edge
		}
	}

	return nil
}

type GraphEdge struct {
	id   string
	from *GraphNodeWidget
	to   *GraphNodeWidget
}

type DraggableGraphWidget struct {
	id                   string
	offset               image.Point
	dragging             bool
	dragModeEdge         bool
	draggingTarget       *GraphNodeWidget
	panning              bool
	lastMousePosition    image.Point
	offsetChanged        func(image.Point)
	nodes                map[string]*GraphNodeWidget
	activeNode           *GraphNodeWidget
	nodePriorities       []*GraphNodeWidget
	nodeCreationFunction func(*DraggableGraphWidget)
	edges                map[string]*GraphEdge
	edgeCreationFunction func(*DraggableGraphWidget, string, string)
	zoom                 float32
}

func DraggableGraph(id string, offsetChanged func(image.Point), nodeCreationMenu func(*DraggableGraphWidget),  edgeCreationMenu func(*DraggableGraphWidget, string, string)) *DraggableGraphWidget {
	return &DraggableGraphWidget{
		id: id,
		offset: image.Point{
			X: 0,
			Y: 0,
		},
		panning:              false,
		dragging:             false,
		dragModeEdge:         false,
		offsetChanged:        offsetChanged,
		nodes:                map[string]*GraphNodeWidget{},
		nodeCreationFunction: nodeCreationMenu,
		edges:                map[string]*GraphEdge{},
		edgeCreationFunction: edgeCreationMenu,
		zoom:                 1.0,
	}
}

func (w *DraggableGraphWidget) GetOffset() image.Point {
	return w.offset
}

func (w *DraggableGraphWidget) CreateNode(id string, position image.Point, nodeColor color.Color, clicked func(), doubleClicked func()) {
	w.nodes[id] = GraphNode(id, position, nodeColor, clicked, doubleClicked)
}

func (w *DraggableGraphWidget) DestroyNode(id string) {
	for _, edge := range w.nodes[id].associatedEdges {
		w.DestroyEdge(edge.id)
	}

	if w.nodes[id] == w.activeNode {
		w.activeNode = nil
	}

	delete(w.nodes, id)
}

func (w *DraggableGraphWidget) CreateEdge(id string, fromId string, toId string) {
	w.edges[id] = &GraphEdge{
		id:   id,
		from: w.nodes[fromId],
		to:   w.nodes[toId],
	}

	w.nodes[fromId].associatedEdges[id] = w.edges[id]
	w.nodes[toId].associatedEdges[id] = w.edges[id]
}

func (w *DraggableGraphWidget) DestroyEdge(id string) {
	if w.edges[id].to != nil {
		delete(w.edges[id].to.associatedEdges, id)
	}

	if w.edges[id].to != nil {
		delete(w.edges[id].from.associatedEdges, id)
	}

	delete(w.edges, id)
}

func (w *DraggableGraphWidget) Build() {
	// Create graph field
	sizeX, sizeY := g.GetAvailableRegion()
	if sizeX <= 0 || sizeY <= 0 {
		return
	}

	g.InvisibleButton().Size(sizeX, sizeY).Build()

	// Calculate usable window position
	wx, wy := g.Context.Backend().GetWindowPos()
	windowPosition := image.Point{int(wx), int(wy)}

	// Calculate the relative position of the mouse
	relativeMousePosition := g.GetMousePos().Sub(windowPosition)

	// Theme nodes
	nodeSize := image.Point{30, 30}

	// Grab our canvas for later
	canvas := g.GetCanvas()

	// Input handing step
	if g.IsMouseDown(g.MouseButtonMiddle) {
		if !w.panning {
			w.panning = true
			w.lastMousePosition = relativeMousePosition
		} else if w.panning {
			w.offset = w.offset.Add(relativeMousePosition.Sub(w.lastMousePosition))
			w.lastMousePosition = relativeMousePosition
		}
	} else if w.panning {
		w.panning = false
	}

	/// Keyboard shortcuts
	if g.IsKeyPressed(g.KeyDelete) && w.activeNode != nil {
		w.DestroyNode(w.activeNode.id)
	}

	if g.IsKeyPressed(g.KeyEnter) && w.activeNode != nil {
		w.activeNode.doubleClicked()
	}

	if g.IsWindowFocused(g.FocusedFlags(g.FocusedFlagsNone)) && g.IsKeyPressed(g.KeyN) {
		w.nodeCreationFunction(w)
	}

	/// Node inputs
	if g.IsMouseDown(g.MouseButtonLeft) {
		if !w.dragging {
			w.dragging = true
			w.lastMousePosition = relativeMousePosition
			w.draggingTarget = findNodeUnderCursor(relativeMousePosition, w.nodePriorities, nodeSize, w.offset)
		} else if w.draggingTarget != nil && !g.IsKeyDown(g.KeyLeftShift) && !w.dragModeEdge {
			w.draggingTarget.position = w.draggingTarget.position.Add(relativeMousePosition.Sub(w.lastMousePosition))
			w.lastMousePosition = relativeMousePosition
		} else if w.draggingTarget != nil && (w.dragModeEdge || g.IsKeyDown(g.KeyLeftShift)) {
			w.dragModeEdge = true
			nodeMiddle := w.draggingTarget.position.Add(nodeSize.Div(2)).Add(windowPosition).Add(w.offset)
			canvas.AddLine(nodeMiddle, g.GetMousePos(), color.White, 1.0)
		}
	} else if w.dragging {
		w.dragging = false

		if g.IsKeyDown(g.KeyLeftShift) || w.dragModeEdge {
			edgeTarget := findNodeUnderCursor(relativeMousePosition, w.nodePriorities, nodeSize, w.offset)

			if edgeTarget != nil {
				edgeToTarget := w.draggingTarget.HasEdgeTo(edgeTarget)

				if edgeToTarget != nil {
					w.DestroyEdge(edgeToTarget.id)
				} else {
					w.edgeCreationFunction(w, w.draggingTarget.id, edgeTarget.id)
					// TODO: Retain a temporary line for drawing
				}
			}
		}

		w.dragModeEdge = false
		w.draggingTarget = nil
	}

	if g.IsMouseDoubleClicked(g.MouseButtonLeft) {
		consumingNode := findNodeUnderCursor(relativeMousePosition, w.nodePriorities, nodeSize, w.offset)
		if consumingNode != nil {
			w.activeNode = consumingNode
			consumingNode.doubleClicked()
		}
	}

	if g.IsMouseClicked(g.MouseButtonLeft) {
		consumingNode := findNodeUnderCursor(relativeMousePosition, w.nodePriorities, nodeSize, w.offset)
		if consumingNode != nil {
			w.activeNode = consumingNode
			consumingNode.clicked()
		} else {
			w.activeNode = nil
		}
	}

	// Draw step
	/// Draw edges
	/// TODO: This is stupid.
	//// Draw gray edges first so that they can be overwritten by white edges
	for _, edge := range w.edges {
		edgeStart := edge.from.position.Add(nodeSize.Div(2)).Add(windowPosition).Add(w.offset)
		edgeEnd := edge.to.position.Add(nodeSize.Div(2)).Add(windowPosition).Add(w.offset)
		edgeMid := edgeStart.Add(edgeEnd.Sub(edgeStart).Div(2))

		canvas.AddLine(edgeMid, edgeEnd, color.Gray{128}, 1.0)
	}

	//// Draw white edges capable of overwriting gray edges
	for _, edge := range w.edges {
		edgeStart := edge.from.position.Add(nodeSize.Div(2)).Add(windowPosition).Add(w.offset)
		edgeEnd := edge.to.position.Add(nodeSize.Div(2)).Add(windowPosition).Add(w.offset)
		edgeMid := edgeStart.Add(edgeEnd.Sub(edgeStart).Div(2))

		canvas.AddLine(edgeStart, edgeMid, color.White, 1.0)
	}

	/// Draw nodes
	newNodePriorities := []*GraphNodeWidget{} // We really should be using a stack here.

	for _, node := range w.nodes {
		nodeRelativePosition := node.position.Add(windowPosition).Add(w.offset)
		if node == w.activeNode {
			canvas.AddRectFilled(nodeRelativePosition, nodeRelativePosition.Add(nodeSize), node.color, 0.0, g.DrawFlagsNone)
		} else {
			canvas.AddRect(nodeRelativePosition, nodeRelativePosition.Add(nodeSize), node.color, 0.0, g.DrawFlagsNone, 1.0)
		}

		newNodePriorities = append(newNodePriorities, node)
	}

	slices.Reverse(newNodePriorities)
	w.nodePriorities = newNodePriorities
}
