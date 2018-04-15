module Main exposing (main)

import Html
import Model exposing (..)
import Task
import Update exposing (..)
import View exposing (..)
import Window


main : Program Never Model Msg
main =
    Html.program
        { init = ( init, Task.perform SetWindowSize Window.size )
        , view = view
        , update = update
        , subscriptions = subscriptions
        }


subscriptions : Model -> Sub Msg
subscriptions model =
    Window.resizes SetWindowSize
