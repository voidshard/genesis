package geography

import (
	"github.com/voidshard/genesis/pkg/types"
)

// Settings are advanced controls for adjusting pretty much everything about the
// geography editor / functions.
//
// Defaults here are reasonable, but creating maps much larger or much smaller
// probably will require tweaking to get things looking right.
type Settings struct {
	// Weights for how much affect each thing has on the final heightmap.
	// These directly multiply vs. uint8 values to establish the final
	// value. You probably want Ravine & River weights to be negative.
	HeightMapMountainWeight     float64
	HeightMapRavineWeight       float64
	HeightMapRiverWeight        float64
	HeightMapNoisePerlinWeight  float64
	HeightMapNoiseVoronoiWeight float64

	// Graph weights for path calculations - these encourage paths
	// to avoid certain points.
	// Eg. GraphEdgeWeight encourages paths to avoid edges.
	// GraphMountainWeight helps mountain ranges avoid each other.
	GraphEdgeWeight     int
	GraphDefaultWeight  int
	GraphMountainWeight int
	GraphRavineWeight   int
	GraphLandWeight     int
	GraphSeaWeight      int

	// Fractal noise settings, used in conjunction with perlin noise so
	// maps have a mix of sharp & round edges
	NoiseFractalSegments   float64
	NoiseFractalIterations int

	// Ocean water temp settings. Since we use uint8 for temperature and
	// want negative values, we consider 100 => 0 degrees C (by default).
	// The poles are very cold, the equator very warm. The rest of the sea
	// is somewhere in the middle.
	//
	// Nb. we reserve '0' here to indicate 'land' so even if you configure
	// a slightly different strategy with respect to temperatures, you should
	// set 'OceanWaterVeryCold' > 0 that 0 can be left for 'not-sea'
	OceanWaterVeryCold uint8 // temperature at poles
	OceanWaterVeryWarm uint8 // temperature at equator
	OceanWaterCold     uint8
	OceanWaterWarm     uint8

	// Ocean current settings; currently we build another highly connected
	// graph for the currents (whose nodes are all in the sea .. obviously).
	OceanCurrentGridSize int
	OceanCurrentWidth    int
	// OceanColdCurrentProb probability of cold water current (vs. hot)
	OceanColdCurrentProb float64

	// Volcano cone & caldera settings.
	// You probably want cone > caldera .. most of the time.
	VolcanoCone       *types.Dice
	VolcanoCaldera    *types.Dice
	VolcanoStep       *types.Dice
	VolcanoRangeWidth int

	// RavineWidth seems .. obvious
	RavineWidth *types.Dice

	// Mountain settings affect size, frequency and range width
	Mountain           *types.Dice
	MountainsPerStep   *types.Dice
	MountainStep       *types.Dice
	MountainRangeWidth int

	// List of prevailing wind directions from the North pole to the South pole,
	// assuming that the world is divided into horizontal bands of equal(ish) size.
	// Used for calculation of rain shadows / desertification etc.
	RainfallPrevailingWinds           []types.Heading
	RainfallStormInitMoisture         *types.Dice
	RainfallMoistureGainVeryWarmSea   *types.Dice
	RainfallMoistureGainWarmSea       *types.Dice
	RainfallMoistureGainColdSea       *types.Dice
	RainfallMoistureGainVeryColdSea   *types.Dice
	RainfallMoistureLossOverLand      *types.Dice
	RainfallMoistureLossOverMountains *types.Dice
	RainfallDryWindMoistureLoss       *types.Dice
	RainfallCalcRoutines              int
}

func DefaultSettings() *Settings {
	return &Settings{
		MountainRangeWidth:          40,
		Mountain:                    types.NewDice(5, 6, 6),
		MountainsPerStep:            types.NewDice(5, 10),
		MountainStep:                types.NewDice(4, 15),
		VolcanoCone:                 types.NewDice(8, 6, 6),
		VolcanoCaldera:              types.NewDice(3, 5),
		VolcanoStep:                 types.NewDice(24, 20),
		VolcanoRangeWidth:           45,
		RavineWidth:                 types.NewDice(2, 10, 10),
		GraphDefaultWeight:          200,
		GraphEdgeWeight:             200,
		GraphMountainWeight:         500,
		GraphRavineWeight:           30,
		GraphLandWeight:             1000,
		GraphSeaWeight:              1000,
		NoiseFractalSegments:        0.25,
		NoiseFractalIterations:      3,
		HeightMapMountainWeight:     0.6,
		HeightMapRavineWeight:       -0.1,
		HeightMapRiverWeight:        0.0,
		HeightMapNoisePerlinWeight:  0.5,
		HeightMapNoiseVoronoiWeight: 0.5,
		OceanWaterVeryCold:          100,
		OceanWaterVeryWarm:          135,
		OceanWaterCold:              105,
		OceanWaterWarm:              125,
		OceanCurrentGridSize:        100,
		OceanCurrentWidth:           30,
		OceanColdCurrentProb:        0.4,
		RainfallPrevailingWinds: []types.Heading{
			types.EAST,      // Polar
			types.WEST,      // Temperate
			types.NORTHEAST, // Tropical
			types.SOUTHWEST, // Tropical
			types.WEST,      // Temperate
			types.EAST,      // Polar
		},
		RainfallStormInitMoisture:         types.NewDice(0),
		RainfallMoistureGainVeryWarmSea:   types.NewDice(10, 10, 5),
		RainfallMoistureGainWarmSea:       types.NewDice(5, 5, 5),
		RainfallMoistureGainColdSea:       types.NewDice(2, 2, 2),
		RainfallMoistureGainVeryColdSea:   types.NewDice(-10, 5),
		RainfallMoistureLossOverLand:      types.NewDice(6, 3, 3, 10),
		RainfallMoistureLossOverMountains: types.NewDice(1, 2, 2),
		RainfallDryWindMoistureLoss:       types.NewDice(1, 2, 2),
		RainfallCalcRoutines:              10,
	}
}
