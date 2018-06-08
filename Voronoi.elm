module Voronoi
    exposing
        ( Cell
        , Diagram
        , Point
        , compute
        )

import Heap


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

        -- impossible, but satisfy the checker.  TODO: crash?
        Just ( VertexEvent _, _ ) ->
            { cells = [] }

        Just ( SiteEvent p, events ) ->
            computeRecurse
                { events = events
                , beach = BeachArc p
                , diagram = { cells = [] }
                }


computeRecurse : State -> Diagram
computeRecurse state =
    case Heap.pop state.events of
        Nothing ->
            state.diagram

        Just ( SiteEvent p, events ) ->
            computeRecurse (State events (insertArc p state.beach) state.diagram)

        Just ( VertexEvent v, events ) ->
            computeRecurse (State events (removeArc v state.beach) state.diagram)


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


type Beach
    = BeachEdge Point Point Beach Beach
    | BeachArc Point


insertArc : Point -> Beach -> Beach
insertArc site beach =
    case beach of
        BeachArc p ->
            BeachEdge
                p
                site
                (BeachArc p)
                (BeachEdge
                    site
                    p
                    (BeachArc site)
                    (BeachArc p)
                )

        BeachEdge pLeft pRight left right ->
            if site.x < edgeX site.y pLeft pRight then
                BeachEdge pLeft pRight (insertArc site left) right
            else
                BeachEdge pLeft pRight left (insertArc site right)


removeArc : Vertex -> Beach -> Beach
removeArc v beach =
    case beach of
        -- impossible, but satisfy the checker.  TODO: crash?
        BeachArc p ->
            BeachArc p

        BeachEdge pLeft pRight left right ->
            if ( pLeft, pRight ) == ( v.left, v.middle ) then
                BeachEdge pLeft v.right left (removeLeftArc right)
            else if ( pLeft, pRight ) == ( v.middle, v.right ) then
                BeachEdge v.left pRight (removeRightArc left) right
            else if v.circumCenter.x < edgeX v.y pLeft pRight then
                BeachEdge pLeft pRight (removeArc v left) right
            else
                BeachEdge pLeft pRight left (removeArc v right)


removeLeftArc : Beach -> Beach
removeLeftArc beach =
    case beach of
        BeachArc p ->
            BeachArc p

        BeachEdge _ _ (BeachArc _) right ->
            right

        BeachEdge pLeft pRight left right ->
            BeachEdge pLeft pRight (removeLeftArc left) right


removeRightArc : Beach -> Beach
removeRightArc beach =
    case beach of
        BeachArc p ->
            BeachArc p

        BeachEdge _ _ left (BeachArc _) ->
            left

        BeachEdge pLeft pRight left right ->
            BeachEdge pLeft pRight left (removeRightArc right)


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
