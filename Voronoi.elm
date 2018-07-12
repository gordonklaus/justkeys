module Voronoi
    exposing
        ( Cell
        , Diagram
        , Point
        , compute
        )

import Heap
import RefMap exposing (RefID, RefMap)


type alias Diagram =
    { cells : List Cell
    }


type alias Cell =
    {}


type alias Point =
    { x : Float
    , y : Float
    }


compute : List Point -> Diagram
compute sites =
    case Heap.pop <| Heap.fromList (Heap.smallest |> Heap.by eventY) (List.map SiteEvent sites) of
        Nothing ->
            { cells = [] }

        Just ( VertexEvent _, _ ) ->
            Debug.crash "unreachable"

        Just ( SiteEvent p, events ) ->
            let
                beachListNode =
                    BeachListNode p Nothing Nothing

                ( beachListNodeID, beachList ) =
                    RefMap.put beachListNode RefMap.empty
            in
            computeRecurse
                { events = events
                , beach = { tree = BeachTreeArc p, list = beachList }
                , diagram = { cells = [] }
                }


computeRecurse : State -> Diagram
computeRecurse state =
    case Heap.pop state.events of
        Nothing ->
            state.diagram

        Just ( SiteEvent p, events ) ->
            let
                ( beach, q ) =
                    insertArc p state.beach
            in
            computeRecurse (State events beach state.diagram)

        Just ( VertexEvent v, events ) ->
            computeRecurse (State events { tree = removeArc v (assertEdge state.beach.tree), list = state.beach.list } state.diagram)


type alias State =
    -- TODO: remove
    { events : Heap.Heap Event
    , beach : Beach
    , diagram : Diagram
    }


type Event
    = SiteEvent Point
    | VertexEvent Vertex


type alias Vertex =
    { y : Float
    , circumCenter : Point
    , left : Point
    , middle : Point
    , right : Point
    }


eventY e =
    case e of
        SiteEvent p ->
            p.y

        VertexEvent v ->
            v.y


type alias Beach =
    { tree : BeachTree
    , list : RefMap BeachListNode
    }


type BeachTree
    = BeachTreeEdge BeachTreeEdgeRecord
    | BeachTreeArc Point


type alias BeachTreeEdgeRecord =
    { pLeft : Point
    , pRight : Point
    , left : BeachTree
    , right : BeachTree
    }


type alias BeachListNode =
    { p : Point
    , left : Maybe BeachListNodeID
    , right : Maybe BeachListNodeID
    }


type BeachListNodeID
    = BeachListNodeID (RefID BeachListNode)


insertArc : Point -> Beach -> ( Beach, Point )
insertArc site beach =
    case beach.tree of
        BeachTreeArc p ->
            let
                tree =
                    BeachTreeEdge
                        { pLeft = p
                        , pRight = site
                        , left = BeachTreeArc p
                        , right =
                            BeachTreeEdge
                                { pLeft = site
                                , pRight = p
                                , left = BeachTreeArc site
                                , right = BeachTreeArc p
                                }
                        }
            in
            ( { beach | tree = tree }, p )

        BeachTreeEdge edge ->
            if site.x < edgeX site.y edge.pLeft edge.pRight then
                applyLeft (insertArc site) edge beach.list
            else
                let
                    ( beach_, split ) =
                        insertArc site { beach | tree = edge.right }
                in
                ( { beach_ | tree = BeachTreeEdge { edge | right = beach_.tree } }, split )


applyLeft : (Beach -> ( Beach, Point )) -> BeachTreeEdgeRecord -> RefMap BeachListNode -> ( Beach, Point )
applyLeft f edge list =
    let
        ( beach_, split ) =
            f { tree = edge.left, list = list }
    in
    ( { beach_ | tree = BeachTreeEdge { edge | left = beach_.tree } }, split )


removeArc : Vertex -> BeachTreeEdgeRecord -> BeachTree
removeArc v edge =
    let
        edge_ =
            if ( edge.pLeft, edge.pRight ) == ( v.left, v.middle ) then
                { edge | pRight = v.right, right = removeLeftArc (assertEdge edge.right) }
            else if ( edge.pLeft, edge.pRight ) == ( v.middle, v.right ) then
                { edge | pLeft = v.left, left = removeRightArc (assertEdge edge.left) }
            else if v.circumCenter.x < edgeX v.y edge.pLeft edge.pRight then
                { edge | left = removeArc v (assertEdge edge.left) }
            else
                { edge | right = removeArc v (assertEdge edge.right) }
    in
    BeachTreeEdge edge_


removeLeftArc : BeachTreeEdgeRecord -> BeachTree
removeLeftArc { left, right } =
    case left of
        BeachTreeArc _ ->
            right

        BeachTreeEdge edge ->
            removeLeftArc edge


removeRightArc : BeachTreeEdgeRecord -> BeachTree
removeRightArc { left, right } =
    case right of
        BeachTreeArc _ ->
            left

        BeachTreeEdge edge ->
            removeRightArc edge


assertEdge : BeachTree -> BeachTreeEdgeRecord
assertEdge beach =
    case beach of
        BeachTreeArc _ ->
            Debug.crash "unreachable"

        BeachTreeEdge edge ->
            edge


edgeX y p1 p2 =
    let
        dx =
            p2.x - p1.x

        dy =
            p2.y - p1.y

        dy1 =
            y - p1.y

        dy2 =
            y - p2.y

        b =
            (p2.x * dy1 - p1.x * dy2) / dy

        b2_c =
            sqrt (dy1 * dy2 * ((dx / dy) ^ 2 + 1))
    in
    if dy < 0 then
        b + b2_c
    else if dy > 0 then
        b - b2_c
    else
        p1.x + dx / 2
