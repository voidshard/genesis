package geography

import (
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/voidshard/genesis/internal/paint"
	"github.com/voidshard/genesis/pkg/types"
)

type rainData struct {
	Start     image.Point
	Area      image.Rectangle
	Direction types.Heading
	Moisture  float64
	Height    uint8
}

func (e *Editor) Rain(proj string, stormMult float64, prevailingWinds []types.Heading) (image.Image, error) {
	p, err := e.project(proj)
	if err != nil {
		return nil, err
	}
	pnt := paint.New(e.cfg.Gen.Root, p.WorldWidth, p.WorldHeight)

	if stormMult <= 0 {
		rain, err := pnt.Canvas(p.Canvas(tagRain))
		return rain.Image(), err
	}

	// we need mountains to know when storms are forced upwards (dumping rain, losing moisture)
	mountains, err := pnt.Canvas(p.Canvas(tagMountains))
	if err != nil {
		return nil, err
	}

	// sea information (sea temperature influences dry / wet winds)
	sea, err := pnt.Canvas(p.Canvas(tagSea))
	if err != nil {
		return nil, err
	}

	// current rain map
	rain, err := pnt.Canvas(p.Canvas(tagRain))
	if err != nil {
		return nil, err
	}

	wg := &sync.WaitGroup{}
	work := make(chan *rainData)

	go func() {
		defer close(work)

		if prevailingWinds == nil || len(prevailingWinds) == 0 {
			prevailingWinds = e.set.RainfallPrevailingWinds
		}

		sliceHeight := p.WorldHeight / len(prevailingWinds)

		for i, direction := range prevailingWinds {
			area := image.Rect(0, i*sliceHeight, p.WorldWidth, (i+1)*sliceHeight)
			for start := range edgePoints(area, direction.Opposite()) {
				work <- &rainData{
					Start:     start,
					Area:      area,
					Direction: direction,
					Moisture:  float64(e.set.RainfallStormInitMoisture.Roll()),
				}
			}
		}
	}()

	for i := 0; i < e.set.RainfallCalcRoutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for data := range work {
				// direction our winds / storm go (pixel dx/dy)
				dx, dy := data.Direction.RiseRun()
				area := data.Area
				storm := data.Start

				for {
					if storm.X < area.Min.X || storm.X > area.Max.X || storm.Y < area.Min.Y || storm.Y >= area.Max.Y {
						break
					}

					// read current values from maps
					rainValue := rain.B(storm.X, storm.Y)
					seaTemp := sea.B(storm.X, storm.Y)
					height := mountains.R(storm.X, storm.Y)

					isLand := seaTemp == 0 // 0 is a reserved value
					if !isLand {           // eg. we're over the sea
						// depending on ocean temp, gain moisture
						if seaTemp >= e.set.OceanWaterVeryWarm {
							data.Moisture += float64(e.set.RainfallMoistureGainVeryWarmSea.Roll()) * stormMult
						} else if seaTemp >= e.set.OceanWaterWarm {
							data.Moisture += float64(e.set.RainfallMoistureGainWarmSea.Roll()) * stormMult
						} else if seaTemp >= e.set.OceanWaterCold {
							data.Moisture += float64(e.set.RainfallMoistureGainColdSea.Roll()) * stormMult
						} else if seaTemp >= e.set.OceanWaterVeryCold {
							data.Moisture += float64(e.set.RainfallMoistureGainVeryColdSea.Roll()) * stormMult
						}
					} else if data.Moisture > 0 { // over land, air contains moisture
						delta := float64(e.set.RainfallMoistureLossOverLand.Roll()) * stormMult

						if height >= 0 && height > data.Height { // going up over mountains
							delta = float64(e.set.RainfallMoistureLossOverMountains.Roll()) * stormMult * float64(height-data.Height)
						}

						if delta >= data.Moisture {
							delta = data.Moisture
						}

						data.Moisture -= delta
						rain.Set(storm.X, storm.Y, color.RGBA{0, 0, incrUint8(rainValue, math.Round(delta)), 255})
					}

					data.Height = height
					storm.X += dx
					storm.Y += dy
				}
			}
		}()
	}

	wg.Wait()

	return rain.Image(), pnt.Save(rain)
}
