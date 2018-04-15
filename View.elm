module View exposing (view)

import Html
import Html.Attributes
import Math.Vector2 exposing (Vec2, vec2)
import Model exposing (..)
import Svg
import Svg.Attributes
import Svg.Events
import Touch
import Update exposing (..)
import Voronoi.Constants
import Voronoi.Delaunay.BowyerWatson
import Voronoi.Geometry.Triangle
import Voronoi.Model
import Voronoi.Voronoi
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


viewVoronoi : List Voronoi.Model.Point -> Svg.Svg Msg
viewVoronoi points =
    let
        tris =
            List.foldl Voronoi.Delaunay.BowyerWatson.addPoint (superTriangles 6 0 4 1) points

        voronoi =
            Voronoi.Voronoi.getVoronoi tris
    in
    Svg.g
        [ Svg.Attributes.name "voronoi"
        , Svg.Attributes.transform "translate(0, .36) scale(1, -1)"
        ]
        (List.map drawVoronoi voronoi)


superTriangles : Float -> Float -> Float -> Float -> List Voronoi.Model.DelaunayTriangle
superTriangles x y width height =
    [ Voronoi.Geometry.Triangle.getDelaunayTriangle
        (Voronoi.Model.Triangle
            (Voronoi.Model.Point (vec2 (x - width) (y - height)))
            (Voronoi.Model.Point (vec2 (x - width) (y + 2 * height)))
            (Voronoi.Model.Point (vec2 (x + 2 * width) (y + 2 * height)))
        )
    , Voronoi.Geometry.Triangle.getDelaunayTriangle
        (Voronoi.Model.Triangle
            (Voronoi.Model.Point (vec2 (x - width) (y - height)))
            (Voronoi.Model.Point (vec2 (x + 2 * width) (y - height)))
            (Voronoi.Model.Point (vec2 (x + 2 * width) (y + 2 * height)))
        )
    ]


drawVoronoi : Voronoi.Model.VoronoiPolygon -> Svg.Svg Msg
drawVoronoi voronoi =
    Svg.polygon
        [ Svg.Attributes.fill "gray"
        , Svg.Attributes.stroke "black"
        , Svg.Attributes.strokeWidth "0.002"
        , Svg.Attributes.points (Voronoi.Voronoi.toString voronoi)
        , Touch.onStart
            (\event ->
                let
                    _ =
                        Debug.log "touch" (Voronoi.Voronoi.toString voronoi)
                in
                None
            )
        ]
        []
