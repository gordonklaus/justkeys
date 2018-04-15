module Update exposing (Msg(..), update)

import Model exposing (..)
import Window


type Msg
    = None
    | SetWindowSize Window.Size


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        None ->
            model ! [ Cmd.none ]

        SetWindowSize size ->
            ( { model | windowSize = size }, Cmd.none )
