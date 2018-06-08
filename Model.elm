module Model exposing (Model, init)

import Voronoi
import Window


type alias Model =
    { windowSize : Window.Size
    , points : List Voronoi.Point
    }


init : Model
init =
    { windowSize = Window.Size 0 0
    , points =
        [ Voronoi.Point 8 0.10522005131820976
        , Voronoi.Point 9 0.06404549305066227
        , Voronoi.Point 9.584962500721156 0.04211434111882985
        , Voronoi.Point 7 0.06404549305066223
        , Voronoi.Point 8.584962500721156 0.17629757010954353
        , Voronoi.Point 9.321928094887362 0.10748616513402995
        , Voronoi.Point 6.415037499278844 0.04211434111883005
        , Voronoi.Point 7.415037499278844 0.17629757010954358
        , Voronoi.Point 8.415037499278844 0.23997001788716987
        , Voronoi.Point 8.736965594166206 0.19990161578833052
        , Voronoi.Point 9.222392421336448 0.15181867493172513
        , Voronoi.Point 9.415037499278844 0.1346633757802257
        , Voronoi.Point 6 0.03147253480439599
        , Voronoi.Point 7.584962500721156 0.23997001788717023
        , Voronoi.Point 8.321928094887362 0.2786280134523594
        , Voronoi.Point 8.807354922057604 0.214980222864391
        , Voronoi.Point 9.169925001442312 0.1659536868562522
        , Voronoi.Point 9.459431618637296 0.1485212766661289
        , Voronoi.Point 5.678071905112638 0.027335231744877488
        , Voronoi.Point 6.678071905112638 0.10748616513402995
        , Voronoi.Point 7.2630344058337934 0.19990161578833054
        , Voronoi.Point 7.678071905112637 0.2786280134523594
        , Voronoi.Point 8.263034405833794 0.28932448742172584
        , Voronoi.Point 8.485426827170242 0.2954150682345159
        , Voronoi.Point 8.678071905112638 0.26085123015698786
        , Voronoi.Point 8.84799690655495 0.22051006105941703
        , Voronoi.Point 9.137503523749935 0.18873997249827687
        , Voronoi.Point 9.263034405833794 0.16574315171270995
        , Voronoi.Point 9.37851162325373 0.1693690401531822
        , Voronoi.Point 9.485426827170242 0.177813942972545
        , Voronoi.Point 7.7369655941662066 0.28932448742172656
        , Voronoi.Point 8.222392421336448 0.3162251985546648
        , Voronoi.Point 8.87446911791614 0.23696934673805453
        , Voronoi.Point 9.115477217419937 0.22643116787037995
        , Voronoi.Point 9.502500340529183 0.18900522031365516
        , Voronoi.Point 6.192645077942396 0.08323728272004807
        , Voronoi.Point 6.777607578663552 0.15181867493172518
        , Voronoi.Point 7.192645077942396 0.214980222864391
        , Voronoi.Point 7.514573172829758 0.29541506823451535
        , Voronoi.Point 7.777607578663552 0.3162251985546648
        , Voronoi.Point 8.192645077942396 0.3230617754680597
        , Voronoi.Point 8.362570079384708 0.3175782603195227
        , Voronoi.Point 8.514573172829758 0.30828907974013203
        , Voronoi.Point 8.652076696579693 0.27981396799766123
        , Voronoi.Point 8.777607578663552 0.23094306574205145
        , Voronoi.Point 8.893084796083489 0.26722278484796647
        , Voronoi.Point 9.099535673550914 0.24738200376274516
        , Voronoi.Point 9.192645077942396 0.1785201769792412
        , Voronoi.Point 9.280107919192735 0.1738511617222514
        , Voronoi.Point 9.362570079384708 0.1696666245136804
        , Voronoi.Point 9.440572591385981 0.1531039215652915
        , Voronoi.Point 9.514573172829758 0.19391751492057122
        , Voronoi.Point 6.584962500721156 0.1346633757802258
        , Voronoi.Point 7.321928094887363 0.26085123015698825
        , Voronoi.Point 7.807354922057604 0.32306177546805986
        , Voronoi.Point 8.169925001442312 0.3326232646563284
        , Voronoi.Point 8.459431618637296 0.30453541699463205
        , Voronoi.Point 8.700439718141093 0.27043858952322086
        , Voronoi.Point 8.906890595608518 0.2829472470452009
        , Voronoi.Point 9.087462841250339 0.26134178039212125
        , Voronoi.Point 9.247927513443585 0.1726248367635938
        , Voronoi.Point 9.39231742277876 0.16240025876206401
        , Voronoi.Point 9.523561956057012 0.19543065425915032
        , Voronoi.Point 6.830074998557688 0.1659536868562522
        , Voronoi.Point 7.15200309344505 0.22051006105941703
        , Voronoi.Point 7.637429920615292 0.3175782603195225
        , Voronoi.Point 7.830074998557688 0.33262326465632847
        , Voronoi.Point 8.15200309344505 0.3482993147587367
        , Voronoi.Point 8.289506617194984 0.3095670308589074
        , Voronoi.Point 8.53051471669878 0.3227831240191221
        , Voronoi.Point 8.637429920615292 0.2977512996807997
        , Voronoi.Point 8.830074998557688 0.23262651584489724
        , Voronoi.Point 8.917537839808027 0.2924077462923238
        , Voronoi.Point 9.078002512001273 0.27094071309459244
        , Voronoi.Point 9.15200309344505 0.18516509843224851
        , Voronoi.Point 9.289506617194984 0.1707342024296936
        , Voronoi.Point 9.3536369546147 0.165555249542973
        , Voronoi.Point 9.473931188332411 0.16724704401382548
        , Voronoi.Point 9.53051471669878 0.19483734264450012
        , Voronoi.Point 6.263034405833794 0.1050382774075266
        , Voronoi.Point 7.485426827170242 0.3082890797401319
        , Voronoi.Point 7.84799690655495 0.3482993147587369
        , Voronoi.Point 8.137503523749935 0.37065493913127645
        , Voronoi.Point 8.37851162325373 0.3271610809990917
        , Voronoi.Point 8.765534746362977 0.23857930194740154
        , Voronoi.Point 8.925999418556223 0.2981468597898507
        , Voronoi.Point 9.070389327891398 0.27734779268942683
        , Voronoi.Point 9.20163386116965 0.17592778630026468
        , Voronoi.Point 9.432959407276106 0.1508217738364088
        , Voronoi.Point 9.53605290024021 0.1929053461891451
        ]
    }