module RefMap
    exposing
        ( RefID
        , RefMap
        , empty
        , get
        , put
        , set
        , update
        )

import Dict


type RefID
    = RefID Int


type RefMap v
    = RefMap { next : Int, dict : Dict.Dict Int v }


empty : RefMap v
empty =
    RefMap { next = 0, dict = Dict.empty }


put : v -> RefMap v -> ( RefID, RefMap v )
put v (RefMap { next, dict }) =
    ( RefID next
    , RefMap { next = next + 1, dict = Dict.insert next v dict }
    )


get : RefID -> RefMap v -> v
get (RefID id) (RefMap { dict }) =
    case Dict.get id dict of
        Nothing ->
            Debug.crash ("no value for key: " ++ toString id)

        Just x ->
            x


set : RefID -> v -> RefMap v -> RefMap v
set (RefID id) v (RefMap m) =
    RefMap { m | dict = Dict.insert id v m.dict }


update : RefID -> (v -> v) -> RefMap v -> RefMap v
update (RefID id) f (RefMap m) =
    RefMap { m | dict = Dict.update id (Maybe.map f) m.dict }
