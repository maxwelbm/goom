package goom

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
)

const debug = false 

const pi = 3.14159
const tau = 2 * pi

const nMapWidth = 16 
const nMapHeight = 16

var fPlayerX = 14.4              
var fPlayerY = 14.7              
var fPlayerA = pi                
const fFOV = pi / 4              
const fDepth = 16.0              
const fSpeed = 9.0               
const fTurnSpeed = fSpeed * 0.75 

const tick = 15 * time.Millisecond

var screen tcell.Screen
var err error

var mazeStyle = tcell.StyleDefault.Background(tcell.ColorBlack).
	Foreground(tcell.ColorDarkSlateBlue)

var skyStyle = tcell.StyleDefault.Background(tcell.ColorBlack).
	Foreground(tcell.ColorWhite)

// Goom is a command line fps
// Based on https://github.com/OneLoneCoder/CommandLineFPS
func Goom() {
	if screen, err = tcell.NewScreen(); err != nil {
		fmt.Println("Failed to start tcell")
		os.Exit(1)
	}
	err = screen.Init()
	if err != nil {
		fmt.Println("Failed to init tcell.Screen")
		os.Exit(1)
	}

	screen.HideCursor()
	screen.SetStyle(skyStyle)
	screen.Clear()

	nScreenWidth, nScreenHeight := screen.Size()

	nSkyHeight := nScreenHeight / 2                    
	fSkyApparentRadius := float64(nScreenWidth) / fFOV 
	nSkyCircumference := int(tau * fSkyApparentRadius) 

	sky := make([][]rune, nSkyCircumference)
	for x := 0; x < nSkyCircumference; x++ {
		sky[x] = make([]rune, nSkyHeight)
		for y := 0; y < nSkyHeight; y++ {
			switch {
			case x == 0 && debug:
				sky[x][y] = '|' 
			case rand.Float64() < 0.02:
				sky[x][y] = '.'
			default:
				sky[x][y] = ' ' 
			}
		}
	}

	const worldMap = "" +
		"#########......." +
		"#..............." +
		"#.......########" +
		"#..............#" +
		"#......##......#" +
		"#......##......#" +
		"#..............#" +
		"###............#" +
		"##.............#" +
		"#......####..###" +
		"#......#.......#" +
		"#......#.......#" +
		"#..............#" +
		"#......#########" +
		"#..............." +
		"################"

	ticker := time.NewTicker(tick)
	tp1 := time.Now()
	tp2 := time.Now()

	for {
		tp2 = time.Now()
		fElapsedTime := tp2.Sub(tp1).Seconds()
		tp1 = tp2

		switch event := screen.PollEvent().(type) {
		case *tcell.EventKey:
			switch {
			case event.Key() == tcell.KeyEscape:
				screen.Fini()
				os.Exit(0)
			case event.Key() == tcell.KeyLeft:
				angle := fPlayerA - fTurnSpeed*tick.Seconds()
				fPlayerA = angle - tau*math.Floor(angle/tau) // mod 2π
			case event.Key() == tcell.KeyRight:
				angle := fPlayerA + fTurnSpeed*tick.Seconds()
				fPlayerA = angle - tau*math.Floor(angle/tau)
			case event.Key() == tcell.KeyUp:
				fPlayerX += math.Sin(fPlayerA) * fSpeed * tick.Seconds()
				fPlayerY += math.Cos(fPlayerA) * fSpeed * tick.Seconds()
				nMapIndex := int(fPlayerX)*nMapWidth + int(fPlayerY)
				if nMapIndex < 0 || nMapIndex >= len(worldMap) || 
					worldMap[nMapIndex] == '#' {
					fPlayerX -= math.Sin(fPlayerA) * fSpeed * tick.Seconds()
					fPlayerY -= math.Cos(fPlayerA) * fSpeed * tick.Seconds()
				}
			case event.Key() == tcell.KeyDown:
				fPlayerX -= math.Sin(fPlayerA) * fSpeed * tick.Seconds()
				fPlayerY -= math.Cos(fPlayerA) * fSpeed * tick.Seconds()
				nMapIndex := int(fPlayerX)*nMapWidth + int(fPlayerY)
				if nMapIndex < 0 || nMapIndex >= len(worldMap) || 
					worldMap[nMapIndex] == '#' {
					fPlayerX += math.Sin(fPlayerA) * fSpeed * tick.Seconds()
					fPlayerY += math.Cos(fPlayerA) * fSpeed * tick.Seconds()
				}
			}
		}

		for x := 0; x < nScreenWidth; x++ {
			fRayAngle := (fPlayerA - fFOV/2.0) + (float64(x) / float64(nScreenWidth) * fFOV)

			fStepSize := 0.1 
			fDistanceToWall := 0.0

			bHitWall := false  
			bBoundary := false 

			fEyeX := math.Sin(fRayAngle) 
			fEyeY := math.Cos(fRayAngle)

			for !bHitWall && fDistanceToWall < fDepth {
				fDistanceToWall += fStepSize
				nTestX := int(fPlayerX + fEyeX*fDistanceToWall)
				nTestY := int(fPlayerY + fEyeY*fDistanceToWall)

				if nTestX < 0 || nTestX >= nMapWidth || nTestY < 0 || nTestY >= nMapHeight {
					bHitWall = true
					fDistanceToWall = fDepth
				} else if worldMap[nTestX*nMapWidth+nTestY] == '#' {
					bHitWall = true 

					p := make([][2]float64, 0)
					for tx := 0; tx < 2; tx++ {
						for ty := 0; ty < 2; ty++ {
							vy := float64(nTestY) + float64(ty) - fPlayerY
							vx := float64(nTestX) + float64(tx) - fPlayerX
							d := math.Sqrt(vx*vx + vy*vy)
							dot := (fEyeX * vx / d) + (fEyeY * vy / d)
							p = append(p, [2]float64{d, dot})
						}
					}

					sort.Slice(p, func(i, j int) bool { return p[i][0] < p[j][0] })

					fBound := 0.01 

					switch {
					case math.Acos(p[0][1]) < fBound:
						bBoundary = p[0][0] < fDistanceToWall
					case math.Acos(p[1][1]) < fBound:
						bBoundary = p[1][0] < fDistanceToWall
					case math.Acos(p[2][1]) < fBound:
						bBoundary = p[2][0] < fDistanceToWall
					}
				}
			}

			nCeiling := float64(nScreenHeight)/2.0 - float64(nScreenHeight)/fDistanceToWall
			nFloor := float64(nScreenHeight) - nCeiling

			var rShade rune 
			switch {
			case bBoundary == true:
				rShade = ' ' 
			case fDistanceToWall <= fDepth/3.0: 
				rShade = '█'
			case fDistanceToWall <= fDepth/2.0:
				rShade = '▓'
			case fDistanceToWall <= fDepth/1.1: 
				rShade = '░'
			default:
				rShade = ' ' 
			}

			for y := 0; y < nScreenHeight; y++ {
				fY := float64(y)
				switch {
				case fY <= nCeiling:
					angle := fPlayerA - pi/8
					angle = angle - tau*math.Floor(angle/tau)
					nPlayerAOffset := (x + int(fSkyApparentRadius*angle)) % nSkyCircumference
					style := skyStyle 
					screen.SetContent(x, y, sky[nPlayerAOffset][y], nil, style)
				case fY > nCeiling && fY <= nFloor:
					screen.SetContent(x, y, rShade, nil, mazeStyle)
				default:
					b := 1.0 - (float64(y)-float64(nScreenHeight)/2.0)/(float64(nScreenHeight)/2.0)
					switch {
					case b < 0.25:
						rShade = '#'
					case b < 0.5:
						rShade = 'x'
					case b < 0.75:
						rShade = '.'
					case b < 0.9:
						rShade = '-'
					default:
						rShade = ' '
					}
					screen.SetContent(x, y, rShade, nil, mazeStyle)
				}
			}
		}
		if debug {
			stats := fmt.Sprintf("X=%3.2f, Y=%3.2f, A=%3.2f, FPS=%3.2f, W=%v, C=%v, R=%v", fPlayerX, fPlayerY, fPlayerA, 1.0/fElapsedTime, nScreenWidth, nSkyCircumference, fSkyApparentRadius)
			for i, c := range stats {
				screen.SetContent(i, 0, c, nil, mazeStyle)
			}
		}

		screen.Show()

		<-ticker.C 
	}
}
