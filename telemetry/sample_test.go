package telemetry

import (
	"math"
	"slices"
	"testing"

	assetpb "github.com/zequent/zqnt-api-go/v3/gen/common/asset/proto"
	livedatapb "github.com/zequent/zqnt-api-go/v3/gen/livedata/proto"
	capabilityv3 "github.com/zequent/zqnt-api-go/v3/gen/zqnt/capability/v3"
	commonv3 "github.com/zequent/zqnt-api-go/v3/gen/zqnt/common/v3"
	telemetryv3 "github.com/zequent/zqnt-api-go/v3/gen/zqnt/telemetry/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var at = &timestamppb.Timestamp{Seconds: 1_791_000_000, Nanos: 250_000_000}

func p[T any](v T) *T { return &v }

func dock() *livedatapb.Telemetry {
	return &livedatapb.Telemetry{
		Id: "sample-1", Timestamp: at, Sn: "DOCK-1",
		Latitude: p(52.5200), Longitude: p(13.4050), AbsoluteAltitude: p[float32](34.5),
		RelativeAltitude: p[float32](0), WindSpeed: p[float32](3.2), Heading: p[float32](181.5),
		Source: &livedatapb.Telemetry_Asset{Asset: &livedatapb.AssetTelemetryDetails{
			EnvironmentTemp: p[float32](21.3), InsideTemp: p[float32](25.1), Humidity: p[float32](48),
			Mode:     p(assetpb.AssetMode_ASSET_MODE_WORKING),
			Rainfall: p(assetpb.RainfallEnum_RAINFALL_LIGHT),
			SubAssetInformation: &livedatapb.AssetTelemetryDetails_AssetSubAssetInformation{
				Sn: p("AIRCRAFT-1"), Model: p("M3TD"), Paired: p(true), Online: p(false)},
			SubAssetAtHome: p(true), SubAssetCharging: p(true), SubAssetPercentage: p[float32](87),
			DebugModeOpen: p(false), HasActiveManualControlSession: p(false),
			CoverState:     p(assetpb.AssetCoverStateEnum_COVER_STATE_CLOSED),
			WorkingVoltage: p[int32](24_000), WorkingCurrent: p[int32](1_250), SupplyVoltage: p[int32](230),
			PositionValid: p(true),
			NetworkInformation: &livedatapb.AssetTelemetryDetails_AssetNetworkInformation{
				Type: p(assetpb.NetworkTypeEnum_NETWORK_TYPE_4_G), Rate: p[float32](512.5),
				Quality: p(assetpb.NetworkStateQualityEnum_NETWORK_STATE_QUALITY_GOOD)},
			AirConditioner: &livedatapb.AssetTelemetryDetails_AssetAirConditioner{
				State: p(assetpb.AssetAirConditionerStateEnum_AIR_CONDITIONER_COOL), SwitchTime: p[int32](30)},
			ManualControlState: p(assetpb.ManualControlStateEnum_MANUAL_CONTROL_STATE_DISCONNECTED),
			PositionState: &livedatapb.AssetTelemetryDetails_PositionState{
				GpsNumber: p[int32](12), RtkNumber: p[int32](20), Quality: p[int32](5)},
			WirelessLink: &livedatapb.AssetTelemetryDetails_AssetWirelessLinkInformation{
				FourthGenerationFreqBand: p[float32](1.8), FourthGenerationGndQuality: p[int32](4),
				FourthGenerationLinkState: p(true), FourthGenerationQuality: p[int32](3),
				FourthGenerationUavQuality: p[int32](2), DongleNumber: p[int32](1), LinkWorkmode: p("SDR_AND_4G"),
				SdrFreqBand: p[float32](5.8), SdrLinkState: p(true), SdrQuality: p[int32](5)},
			SdrState: &livedatapb.AssetTelemetryDetails_AssetSdrState{
				DownQuality: p[int32](5), UpQuality: p[int32](4), FrequencyBand: p(2.4)},
		}},
	}
}

func drone() *livedatapb.Telemetry {
	return &livedatapb.Telemetry{
		Id: "sample-2", Timestamp: at, Sn: "AIRCRAFT-1",
		Latitude: p(52.5210), Longitude: p(13.4070), AbsoluteAltitude: p[float32](110.2),
		RelativeAltitude: p[float32](75.7), WindSpeed: p[float32](5.1), Heading: p[float32](-12.25),
		Source: &livedatapb.Telemetry_SubAsset{SubAsset: &livedatapb.SubAssetTelemetryDetails{
			HorizontalSpeed: p[float32](12.3), VerticalSpeed: p[float32](-0.4), WindDirection: p("NORTH_EAST"),
			Gear: p[int32](1), HeightLimit: p[int32](120), HomeDistance: p[float32](842.6),
			TotalMovementDistance: p(123_456.7), TotalMovementTime: p[float64](98_765),
			Mode: p(assetpb.SubAssetMode_SUBASSET_MODE_WAYLINE), Country: p("DE"),
			BatteryInformation: &livedatapb.SubAssetTelemetryDetails_SubAssetBatteryInformation{
				Percentage: p("64"), RemainingTime: p[int32](1_320), ReturnToHomePower: p("22")},
			PayloadTelemetry: &livedatapb.PayloadTelemetry{
				Id: "payload-0", Name: "H20T", Timestamp: at,
				CameraData: &livedatapb.PayloadTelemetry_CameraData{
					CurrentLens: p("ir"), GimbalPitch: p[float32](-45.5), GimbalYaw: p[float32](10),
					ZoomFactor: p[float32](2), GimbalRoll: p[float32](0.1)},
				RangeFinderData: &livedatapb.PayloadTelemetry_RangeFinderData{
					TargetLatitude: p(52.53), TargetLongitude: p(13.41),
					TargetDistance: p[float32](150.5), TargetAltitude: p[float32](30)},
				SensorData: &livedatapb.PayloadTelemetry_SensorData{TargetTemperature: p[float32](36.6)},
			},
		}},
	}
}

func withoutID(t *livedatapb.Telemetry) *livedatapb.Telemetry {
	c := proto.Clone(t).(*livedatapb.Telemetry)
	c.Id = ""
	return c
}

func TestADockSampleSurvivesTheRoundTrip(t *testing.T) {
	if got := ToTelemetry(ToSample(dock(), "")); !proto.Equal(withoutID(dock()), got) {
		t.Fatalf("round trip changed the dock:\nwant %v\ngot  %v", withoutID(dock()), got)
	}
}

func TestADroneSampleSurvivesTheRoundTrip(t *testing.T) {
	if got := ToTelemetry(ToSample(drone(), "")); !proto.Equal(withoutID(drone()), got) {
		t.Fatalf("round trip changed the drone:\nwant %v\ngot  %v", withoutID(drone()), got)
	}
}

func TestTheSharedFieldsAreTyped(t *testing.T) {
	d := ToSample(dock(), "")
	if d.GetAsset().GetSn() != "DOCK-1" || !proto.Equal(d.GetObservedAt(), at) {
		t.Fatalf("asset/observed_at: %v %v", d.GetAsset(), d.GetObservedAt())
	}
	want := &commonv3.GeoPoint{Latitude: 52.52, Longitude: 13.405, Altitude: p(34.5)}
	if !proto.Equal(want, d.GetPosition()) {
		t.Fatalf("position: %v", d.GetPosition())
	}
	if d.GetHeadingDegrees() != 181.5 {
		t.Fatalf("heading: %v", d.GetHeadingDegrees())
	}
	if d.GetBatteryPercent() != 87 {
		t.Fatalf("a dock's battery is its aircraft's, as the console shows it: %v", d.GetBatteryPercent())
	}
	if d.HorizontalSpeed != nil {
		t.Fatal("a dock does not move")
	}

	a := ToSample(drone(), "")
	if a.GetRelativeAltitude() != 75.7 || a.GetHorizontalSpeed() != 12.3 || a.GetVerticalSpeed() != -0.4 || a.GetBatteryPercent() != 64 {
		t.Fatalf("aircraft: %v", a)
	}
}

func TestDetailsUseTheDottedKeysAndEnumNamesWithoutPrefix(t *testing.T) {
	d := ToSample(dock(), "").GetDetails().GetFields()
	for key, want := range map[string]string{
		DockCoverState: "CLOSED", DockMode: "WORKING", DockAirConditionerState: "COOL",
		DockRainfall: "LIGHT", NetworkType: "4_G", NetworkQuality: "GOOD",
	} {
		if got := d[key].GetStringValue(); got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
	for key, want := range map[string]float64{DockEnvironmentTemperature: 21.3, LinkSdrQuality: 5, WindSpeed: 3.2} {
		if got := d[key].GetNumberValue(); got != want {
			t.Errorf("%s = %v, want %v", key, got, want)
		}
	}
	if !d[DockDroneAtHome].GetBoolValue() {
		t.Error("dock.drone.at_home")
	}

	a := ToSample(drone(), "").GetDetails().GetFields()
	if a[DroneMode].GetStringValue() != "WAYLINE" || a[CameraGimbalPitch].GetNumberValue() != -45.5 {
		t.Errorf("aircraft details: %v", a)
	}
	if got := a[PayloadObservedAt].GetStringValue(); got != "2026-10-03T04:00:00.250Z" {
		t.Errorf("payload.observed_at = %q", got)
	}
}

func TestEveryDetailKeyIsDescribedOnceWithItsType(t *testing.T) {
	described := map[string]*capabilityv3.TelemetryField{}
	for _, f := range Fields() {
		if described[f.GetKey()] != nil {
			t.Fatalf("%s described twice", f.GetKey())
		}
		described[f.GetKey()] = f
	}
	for _, sample := range []*livedatapb.Telemetry{dock(), drone()} {
		for key := range ToSample(sample, "").GetDetails().GetFields() {
			if described[key] == nil {
				t.Errorf("%s is written but not described", key)
			}
		}
	}

	cover := described[DockCoverState]
	if cover.GetType() != capabilityv3.TelemetryValueType_TELEMETRY_VALUE_TYPE_STRING ||
		!slices.Equal(cover.GetAllowedValues(), []string{"CLOSED", "OPENED", "HALF_OPEN", "ABNORMAL"}) {
		t.Errorf("cover state: %v", cover)
	}
	temp := described[DockEnvironmentTemperature]
	if temp.GetType() != capabilityv3.TelemetryValueType_TELEMETRY_VALUE_TYPE_NUMBER || temp.GetUnit() != "°C" {
		t.Errorf("environment temperature: %v", temp)
	}
	if described[DockDroneCharging].GetType() != capabilityv3.TelemetryValueType_TELEMETRY_VALUE_TYPE_BOOLEAN {
		t.Error("dock.drone.charging is a boolean")
	}
	if !proto.Equal(Field(DockCoverState), cover) || Field("radar.mode") != nil {
		t.Error("Field looks a key up in the catalog")
	}
}

func TestKeysV2HasNoFieldForAreDroppedThereAndTheSourceFollowsTheKeys(t *testing.T) {
	sample := &telemetryv3.TelemetrySample{
		Asset: &commonv3.AssetRef{Sn: "RADAR-1"}, ObservedAt: at, BatteryPercent: p(55.5),
		Details: &structpb.Struct{Fields: map[string]*structpb.Value{
			"radar.mode":   structpb.NewStringValue("SCAN"),
			DockCoverState: structpb.NewStringValue("OPENED"),
			DockHumidity:   structpb.NewStringValue("not a number"),
		}},
	}
	dockFrame := ToTelemetry(sample).GetAsset()
	if dockFrame == nil || dockFrame.GetCoverState() != assetpb.AssetCoverStateEnum_COVER_STATE_OPENED {
		t.Fatalf("dock frame: %v", dockFrame)
	}
	if dockFrame.Humidity != nil {
		t.Error("a value of the wrong type is not guessed at")
	}
	if dockFrame.GetSubAssetPercentage() != 55.5 {
		t.Errorf("battery: %v", dockFrame.GetSubAssetPercentage())
	}

	moving := proto.Clone(sample).(*telemetryv3.TelemetrySample)
	moving.HorizontalSpeed = p[float64](4)
	aircraft := ToTelemetry(moving).GetSubAsset()
	if aircraft == nil {
		t.Fatal("a moving asset is an aircraft; the dock's keys are dropped")
	}
	if got := aircraft.GetBatteryInformation().GetPercentage(); got != "55.5" {
		t.Errorf("battery: %q", got)
	}
}

func TestASampleBecomesTheFrameAnAdapterWouldHaveStreamed(t *testing.T) {
	sample := ToSample(drone(), "")
	sample.Asset = &commonv3.AssetRef{Sn: "AIRCRAFT-1", Id: "aircraft-1-id"}
	req := ToRequest(sample)
	if req.GetBase().GetSn() != "AIRCRAFT-1" || req.GetBase().GetAssetId() != "aircraft-1-id" ||
		!proto.Equal(req.GetBase().GetTimestamp(), at) {
		t.Fatalf("base: %v", req.GetBase())
	}
	if req.GetBase().GetTid() == "" || req.GetData().GetId() == "" || req.GetBase().GetTid() == req.GetData().GetId() {
		t.Fatalf("ids: %q %q", req.GetBase().GetTid(), req.GetData().GetId())
	}
	back := RequestToSample(req)
	back.Asset = sample.GetAsset()
	if !proto.Equal(sample, back) {
		t.Fatalf("frame changed the sample:\nwant %v\ngot  %v", sample, back)
	}
}

func TestAFrameWithoutCoordinatesOrSerialStillMaps(t *testing.T) {
	sample := ToSample(&livedatapb.Telemetry{Latitude: p(math.NaN()), Longitude: p(math.NaN())}, "FROM-BASE")
	if sample.Position != nil {
		t.Error("omitted coordinates arrive as NaN")
	}
	if sample.GetAsset().GetSn() != "FROM-BASE" {
		t.Errorf("sn: %q", sample.GetAsset().GetSn())
	}
	if RequestToSample(&livedatapb.ProduceTelemetryRequest{}) != nil {
		t.Error("a frame without data has no sample")
	}
}

func TestTheCatalogMatchesTheJavaOne(t *testing.T) {
	if got := len(Fields()); got != 64 {
		t.Fatalf("%d keys; zqnt-utils-java's TelemetrySampleMapper describes 64", got)
	}
}
