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
                , beach = { tree = BeachTreeArc beachListNodeID, list = { nodes = beachList } }
                , diagram = { cells = [] }
                }


computeRecurse : State -> Diagram
computeRecurse state =
    case Heap.pop state.events of
        Nothing ->
            state.diagram

        Just ( SiteEvent p, events ) ->
            let
                ( beach, qID ) =
                    insertArc p state.beach

                q =
                    RefMap.get qID beach.list.nodes

                getRef =
                    flip RefMap.get beach.list.nodes

                left =
                    Maybe.map getRef q.left

                right =
                    Maybe.map getRef q.right

                events2 =
                    [ Maybe.map2 (\left right -> deleteEvent ( left.p, q.p, right.p )) left right
                    , Maybe.map (\left -> pushEvent ( left, q.p, p )) left
                    , Maybe.map (\right -> pushEvent ( p, q.p, right )) right
                    ]
                        |> List.filterMap identity
                        |> List.foldl (<|) events
            in
            computeRecurse (State events2 beach state.diagram)

        Just ( VertexEvent v, events ) ->
            computeRecurse (State events { tree = removeArc v (assertEdge state.beach.tree), list = state.beach.list } state.diagram)


deleteEvent e events =
    events


pushEvent e events =
    events


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
    , list : BeachList
    }


type BeachTree
    = BeachTreeEdge BeachTreeEdgeRecord
    | BeachTreeArc BeachListNodeID


type alias BeachTreeEdgeRecord =
    { pLeft : Point
    , pRight : Point
    , left : BeachTree
    , right : BeachTree
    }


type alias BeachList =
    { nodes : RefMap BeachListNode }


type alias BeachListNode =
    { p : Point
    , left : Maybe BeachListNodeID
    , right : Maybe BeachListNodeID
    }


type alias BeachListNodeID =
    RefID


insertArc : Point -> Beach -> ( Beach, BeachListNodeID )
insertArc site beach =
    case beach.tree of
        BeachTreeArc pID ->
            let
                old =
                    RefMap.get pID beach.list.nodes

                ( newID, list2 ) =
                    replace pID site beach.list

                ( leftID, list3 ) =
                    insertLeft newID old.p list2

                ( rightID, list4 ) =
                    insertRight newID old.p list3

                tree =
                    BeachTreeEdge
                        { pLeft = old.p
                        , pRight = site
                        , left = BeachTreeArc leftID
                        , right =
                            BeachTreeEdge
                                { pLeft = site
                                , pRight = old.p
                                , left = BeachTreeArc newID
                                , right = BeachTreeArc rightID
                                }
                        }
            in
            ( { tree = tree, list = list4 }, pID )

        BeachTreeEdge edge ->
            if site.x < edgeX site.y edge.pLeft edge.pRight then
                applyLeft (insertArc site) edge beach.list
            else
                let
                    ( beach_, split ) =
                        insertArc site { beach | tree = edge.right }
                in
                ( { beach_ | tree = BeachTreeEdge { edge | right = beach_.tree } }, split )


replace : BeachListNodeID -> Point -> BeachList -> ( BeachListNodeID, BeachList )
replace id p { nodes } =
    let
        old =
            RefMap.get id nodes

        ( newID, nodes2 ) =
            RefMap.put { old | p = p } nodes

        list =
            setRight old.left newID { nodes = nodes2 }

        list2 =
            setLeft old.right newID list

        -- TODO: delete old
    in
    ( newID, list2 )


insertLeft : BeachListNodeID -> Point -> BeachList -> ( BeachListNodeID, BeachList )
insertLeft id p { nodes } =
    let
        left =
            (RefMap.get id nodes).left

        ( newID, nodes2 ) =
            RefMap.put { p = p, left = left, right = Just id } nodes

        list =
            setRight left newID { nodes = nodes2 }

        list2 =
            setLeft (Just id) newID list
    in
    ( newID, list2 )


insertRight : BeachListNodeID -> Point -> BeachList -> ( BeachListNodeID, BeachList )
insertRight id p { nodes } =
    let
        right =
            (RefMap.get id nodes).right

        ( newID, nodes2 ) =
            RefMap.put { p = p, left = Just id, right = right } nodes

        list =
            setLeft right newID { nodes = nodes2 }

        list2 =
            setRight (Just id) newID list
    in
    ( newID, list2 )


setLeft : Maybe BeachListNodeID -> BeachListNodeID -> BeachList -> BeachList
setLeft id left { nodes } =
    let
        nodes2 =
            Maybe.withDefault identity (Maybe.map (\id -> RefMap.update id (\n -> { n | left = Just left })) id) nodes
    in
    { nodes = nodes2 }


setRight : Maybe BeachListNodeID -> BeachListNodeID -> BeachList -> BeachList
setRight id right { nodes } =
    let
        nodes2 =
            Maybe.withDefault identity (Maybe.map (\id -> RefMap.update id (\n -> { n | right = Just right })) id) nodes
    in
    { nodes = nodes2 }


applyLeft : (Beach -> ( Beach, BeachListNodeID )) -> BeachTreeEdgeRecord -> BeachList -> ( Beach, BeachListNodeID )
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
