module Voronoi.Geometry.Point exposing (..)

import Color
import Math.Vector2 exposing (getX, getY, vec2)
import Svg exposing (Svg, circle, line, rect)
import Svg.Attributes exposing (..)
import Voronoi.ColorHelper exposing (colorToHex)
import Voronoi.Model exposing (Model, Point)


-- View
-- Controller


roundPoint : Point -> Point
roundPoint point =
    Point
        (vec2 (Basics.toFloat (round (getX point.pos)))
            (Basics.toFloat (round (getY point.pos)))
        )


toString : Point -> String
toString point =
    String.concat
        (List.intersperse
            ","
            [ Basics.toString (getX point.pos)
            , Basics.toString (getY point.pos)
            ]
        )
