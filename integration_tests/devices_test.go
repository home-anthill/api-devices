package integration_tests

//var _ = Describe("Devices", func() {
//var ctx context.Context
//var logger *zap.SugaredLogger
//var router *gin.Engine
//var collProfiles *mongo.Collection
//var collHomes *mongo.Collection
//var collDevices *mongo.Collection
//
//var currDate = time.Now()
//var deviceController = models.Device{
//	ID:           primitive.NewObjectID(),
//	Mac:          "11:22:33:44:55:66",
//	Manufacturer: "test",
//	Model:        "test",
//	UUID:         uuid.NewString(),
//	Features: []models.Feature{{
//		UUID:   uuid.NewString(),
//		Type:   "controller",
//		Name:   "ac-beko",
//		Enable: true,
//		Order:  1,
//		Unit:   "-",
//	}},
//	CreatedAt:  currDate,
//	ModifiedAt: currDate,
//}
//var deviceSensor = models.Device{
//	ID:           primitive.NewObjectID(),
//	Mac:          "AA:22:33:44:55:BB",
//	Manufacturer: "test2",
//	Model:        "test2",
//	UUID:         uuid.NewString(),
//	Features: []models.Feature{{
//		UUID:   uuid.NewString(),
//		Type:   "sensor",
//		Name:   "temperature",
//		Enable: true,
//		Order:  1,
//		Unit:   "°C",
//	}, {
//		UUID:   uuid.NewString(),
//		Type:   "sensor",
//		Name:   "light",
//		Enable: true,
//		Order:  1,
//		Unit:   "lux",
//	}},
//	CreatedAt:  currDate,
//	ModifiedAt: currDate,
//}
//var home = models.Home{
//	ID:       primitive.NewObjectID(),
//	Name:     "home1",
//	Location: "location1",
//	Rooms: []models.Room{{
//		ID:         primitive.NewObjectID(),
//		Name:       "room1",
//		Floor:      1,
//		CreatedAt:  currDate,
//		ModifiedAt: currDate,
//		Devices:    []primitive.ObjectID{},
//	}},
//	CreatedAt:  currDate,
//	ModifiedAt: currDate,
//}
//
//	BeforeEach(func() {
//		// 1. Init config
//		logger = init_config.BuildConfig()
//		defer logger.Sync()
//
//		err := os.Setenv("SINGLE_USER_LOGIN_EMAIL", "test@test.com")
//		Expect(err).ShouldNot(HaveOccurred())
//
//		// 2. Init server
//		port := os.Getenv("HTTP_PORT")
//		httpOrigin := os.Getenv("HTTP_SERVER") + ":" + port
//
//		router, ctx, collProfiles, collHomes, collDevices = init_config.BuildServer(httpOrigin, logger)
//	})
//
//	AfterEach(func() {
//		test_utils.DropAllCollections(ctx, collProfiles, collHomes, collDevices)
//	})
//})
