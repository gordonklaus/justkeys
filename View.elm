module View exposing (view)

import Html
import Html.Attributes
import Model exposing (..)
import Svg
import Svg.Attributes
import Svg.Events
import Touch
import Update exposing (..)
import Voronoi
import Window


view : Model -> Html.Html Msg
view model =
    Html.div
        []
        [ Svg.svg
            [ Svg.Attributes.width (toString model.windowSize.width)
            , Svg.Attributes.height (toString model.windowSize.height)
            , Svg.Attributes.viewBox "8 0 1.58 .36"
            , Svg.Attributes.preserveAspectRatio "none"
            ]
            [ viewVoronoi model.points
            ]
        ]


viewVoronoi : List Voronoi.Point -> Svg.Svg Msg
viewVoronoi points =
    let
        voronoi =
            Voronoi.compute points
    in
    Svg.g
        [ Svg.Attributes.name "voronoi"
        , Svg.Attributes.transform "translate(0, .36) scale(1, -1)"
        ]
        (List.map drawVoronoiCell voronoi.cells)


drawVoronoiCell : Voronoi.Cell -> Svg.Svg Msg
drawVoronoiCell cell =
    Svg.polygon
        [ Svg.Attributes.fill "gray"
        , Svg.Attributes.stroke "black"
        , Svg.Attributes.strokeWidth "0.002"
        , Svg.Attributes.points ""
        , Touch.onStart
            (\event ->
                let
                    _ =
                        Debug.log "touch" ""
                in
                None
            )
        ]
        []
