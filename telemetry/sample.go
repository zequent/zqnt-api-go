// Package telemetry maps v2 Telemetry and v3 TelemetrySample both ways. The fields every asset
// shares are typed on the sample; every other v2 field is a details entry under one of the keys
// below, enum values as their v2 name without the enum's prefix. Fields describes those keys for
// CapabilitySet.telemetry_fields.
//
// Toward v2 a sample is a dock's (asset) or an aircraft's (sub_asset) by the keys it carries; keys
// of the other kind, and keys v2 has no field for, are dropped there. Telemetry.id and
// SubAssetTelemetryDetails.component_telemetry are not carried.
//
// The keys and the v2 fields behind them are the same as com.zqnt.utils.telemetry.TelemetrySampleMapper
// in zqnt-utils-java.
package telemetry

import (
	"crypto/rand"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	commonpb "github.com/zequent/zqnt-api-go/gen/common/base/proto"
	livedatapb "github.com/zequent/zqnt-api-go/gen/livedata/proto"
	capabilityv3 "github.com/zequent/zqnt-api-go/gen/zqnt/capability/v3"
	commonv3 "github.com/zequent/zqnt-api-go/gen/zqnt/common/v3"
	telemetryv3 "github.com/zequent/zqnt-api-go/gen/zqnt/telemetry/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	WindSpeed     = "wind.speed"
	WindDirection = "wind.direction"

	DockEnvironmentTemperature     = "dock.environment_temperature"
	DockInsideTemperature          = "dock.inside_temperature"
	DockHumidity                   = "dock.humidity"
	DockMode                       = "dock.mode"
	DockRainfall                   = "dock.rainfall"
	DockCoverState                 = "dock.cover_state"
	DockDebugModeOpen              = "dock.debug_mode_open"
	DockPositionValid              = "dock.position_valid"
	DockWorkingVoltage             = "dock.working_voltage"
	DockWorkingCurrent             = "dock.working_current"
	DockSupplyVoltage              = "dock.supply_voltage"
	DockAirConditionerState        = "dock.air_conditioner.state"
	DockAirConditionerSwitchTime   = "dock.air_conditioner.switch_time"
	DockManualControlState         = "dock.manual_control.state"
	DockManualControlActiveSession = "dock.manual_control.active_session"
	DockDroneSN                    = "dock.drone.sn"
	DockDroneModel                 = "dock.drone.model"
	DockDronePaired                = "dock.drone.paired"
	DockDroneOnline                = "dock.drone.online"
	DockDroneAtHome                = "dock.drone.at_home"
	DockDroneCharging              = "dock.drone.charging"
	DockGnssGpsSatellites          = "dock.gnss.gps_satellites"
	DockGnssRtkSatellites          = "dock.gnss.rtk_satellites"
	DockGnssQuality                = "dock.gnss.quality"
	NetworkType                    = "network.type"
	NetworkRate                    = "network.rate"
	NetworkQuality                 = "network.quality"
	Link4gFrequencyBand            = "link.4g.frequency_band"
	Link4gGroundQuality            = "link.4g.ground_quality"
	Link4gLinkState                = "link.4g.link_state"
	Link4gQuality                  = "link.4g.quality"
	Link4gAircraftQuality          = "link.4g.aircraft_quality"
	LinkDongleCount                = "link.dongle_count"
	LinkWorkMode                   = "link.work_mode"
	LinkSdrFrequencyBand           = "link.sdr.frequency_band"
	LinkSdrLinkState               = "link.sdr.link_state"
	LinkSdrQuality                 = "link.sdr.quality"
	SdrDownQuality                 = "sdr.down_quality"
	SdrUpQuality                   = "sdr.up_quality"
	SdrFrequencyBand               = "sdr.frequency_band"
	DroneMode                      = "drone.mode"
	DroneGear                      = "drone.gear"
	DroneHeightLimit               = "drone.height_limit"
	DroneHomeDistance              = "drone.home_distance"
	DroneTotalFlightDistance       = "drone.total_flight_distance"
	DroneTotalFlightTime           = "drone.total_flight_time"
	DroneCountry                   = "drone.country"
	DroneBatteryRemainingTime      = "drone.battery.remaining_time"
	DroneBatteryReturnToHomePower  = "drone.battery.return_to_home_power"
	PayloadID                      = "payload.id"
	PayloadName                    = "payload.name"
	PayloadObservedAt              = "payload.observed_at"
	CameraCurrentLens              = "camera.current_lens"
	CameraGimbalPitch              = "camera.gimbal_pitch"
	CameraGimbalYaw                = "camera.gimbal_yaw"
	CameraGimbalRoll               = "camera.gimbal_roll"
	CameraZoomFactor               = "camera.zoom_factor"
	RangeFinderTargetLatitude      = "range_finder.target_latitude"
	RangeFinderTargetLongitude     = "range_finder.target_longitude"
	RangeFinderTargetDistance      = "range_finder.target_distance"
	RangeFinderTargetAltitude      = "range_finder.target_altitude"
	SensorTargetTemperature        = "sensor.target_temperature"
)

type source int

const (
	sourceNone source = iota
	sourceAsset
	sourceSubAsset
)

type detail struct {
	key         string
	path        []protoreflect.FieldDescriptor
	source      source
	unit        string
	description string
}

var details = []detail{
	newDetail(WindSpeed, "wind_speed", "m/s", "Wind speed"),
	newDetail(DockEnvironmentTemperature, "asset.environment_temp", "°C", "Temperature outside the dock"),
	newDetail(DockInsideTemperature, "asset.inside_temp", "°C", "Temperature inside the dock"),
	newDetail(DockHumidity, "asset.humidity", "%", "Relative humidity inside the dock"),
	newDetail(DockMode, "asset.mode", "", "Operating mode of the dock"),
	newDetail(DockRainfall, "asset.rainfall", "", "Rainfall at the dock"),
	newDetail(DockCoverState, "asset.cover_state", "", "State of the dock cover"),
	newDetail(DockDebugModeOpen, "asset.debug_mode_open", "", "Whether remote debugging is open"),
	newDetail(DockPositionValid, "asset.position_valid", "", "Whether the dock's position is calibrated"),
	newDetail(DockWorkingVoltage, "asset.working_voltage", "mV", "Working voltage"),
	newDetail(DockWorkingCurrent, "asset.working_current", "mA", "Working current"),
	newDetail(DockSupplyVoltage, "asset.supply_voltage", "V", "Supply voltage"),
	newDetail(DockAirConditionerState, "asset.air_conditioner.state", "", "Air conditioner state"),
	newDetail(DockAirConditionerSwitchTime, "asset.air_conditioner.switch_time", "s",
		"Time until the air conditioner may switch mode again"),
	newDetail(DockManualControlState, "asset.manual_control_state", "", "Manual control link state"),
	newDetail(DockManualControlActiveSession, "asset.has_active_manual_control_session", "",
		"Whether a manual control session is active"),
	newDetail(DockDroneSN, "asset.sub_asset_information.sn", "", "Serial of the docked aircraft"),
	newDetail(DockDroneModel, "asset.sub_asset_information.model", "", "Model of the docked aircraft"),
	newDetail(DockDronePaired, "asset.sub_asset_information.paired", "", "Whether the aircraft is paired"),
	newDetail(DockDroneOnline, "asset.sub_asset_information.online", "", "Whether the aircraft is online"),
	newDetail(DockDroneAtHome, "asset.sub_asset_at_home", "", "Whether the aircraft is in the dock"),
	newDetail(DockDroneCharging, "asset.sub_asset_charging", "", "Whether the aircraft is charging"),
	newDetail(DockGnssGpsSatellites, "asset.position_state.gps_number", "", "GPS satellites in view"),
	newDetail(DockGnssRtkSatellites, "asset.position_state.rtk_number", "", "RTK satellites in view"),
	newDetail(DockGnssQuality, "asset.position_state.quality", "", "Positioning quality level"),
	newDetail(NetworkType, "asset.network_information.type", "", "Uplink network type"),
	newDetail(NetworkRate, "asset.network_information.rate", "KB/s", "Uplink network rate"),
	newDetail(NetworkQuality, "asset.network_information.quality", "", "Uplink network quality"),
	newDetail(Link4gFrequencyBand, "asset.wireless_link.fourth_generation_freq_band", "", "4G link frequency band"),
	newDetail(Link4gGroundQuality, "asset.wireless_link.fourth_generation_gnd_quality", "", "4G link quality, ground side"),
	newDetail(Link4gLinkState, "asset.wireless_link.fourth_generation_link_state", "", "Whether the 4G link is up"),
	newDetail(Link4gQuality, "asset.wireless_link.fourth_generation_quality", "", "4G link quality"),
	newDetail(Link4gAircraftQuality, "asset.wireless_link.fourth_generation_uav_quality", "", "4G link quality, aircraft side"),
	newDetail(LinkDongleCount, "asset.wireless_link.dongle_number", "", "4G dongles in use"),
	newDetail(LinkWorkMode, "asset.wireless_link.link_workmode", "", "Wireless link work mode"),
	newDetail(LinkSdrFrequencyBand, "asset.wireless_link.sdr_freq_band", "GHz", "SDR link frequency band"),
	newDetail(LinkSdrLinkState, "asset.wireless_link.sdr_link_state", "", "Whether the SDR link is up"),
	newDetail(LinkSdrQuality, "asset.wireless_link.sdr_quality", "", "SDR link quality"),
	newDetail(SdrDownQuality, "asset.sdr_state.down_quality", "", "SDR downlink quality"),
	newDetail(SdrUpQuality, "asset.sdr_state.up_quality", "", "SDR uplink quality"),
	newDetail(SdrFrequencyBand, "asset.sdr_state.frequency_band", "GHz", "SDR frequency band"),
	newDetail(WindDirection, "sub_asset.wind_direction", "", "Wind direction"),
	newDetail(DroneMode, "sub_asset.mode", "", "Flight mode of the aircraft"),
	newDetail(DroneGear, "sub_asset.gear", "", "Flight gear"),
	newDetail(DroneHeightLimit, "sub_asset.height_limit", "m", "Height limit"),
	newDetail(DroneHomeDistance, "sub_asset.home_distance", "m", "Distance to the home point"),
	newDetail(DroneTotalFlightDistance, "sub_asset.total_movement_distance", "m", "Total flight distance"),
	newDetail(DroneTotalFlightTime, "sub_asset.total_movement_time", "s", "Total flight time"),
	newDetail(DroneCountry, "sub_asset.country", "", "Country the aircraft is in"),
	newDetail(DroneBatteryRemainingTime, "sub_asset.battery_information.remaining_time", "s",
		"Remaining flight time on the battery"),
	newDetail(DroneBatteryReturnToHomePower, "sub_asset.battery_information.return_to_home_power", "",
		"Battery needed to return home"),
	newDetail(PayloadID, "sub_asset.payload_telemetry.id", "", "Payload id"),
	newDetail(PayloadName, "sub_asset.payload_telemetry.name", "", "Payload name"),
	newDetail(PayloadObservedAt, "sub_asset.payload_telemetry.timestamp", "", "When the payload reported (RFC 3339)"),
	newDetail(CameraCurrentLens, "sub_asset.payload_telemetry.camera_data.current_lens", "", "Active camera lens"),
	newDetail(CameraGimbalPitch, "sub_asset.payload_telemetry.camera_data.gimbal_pitch", "°", "Gimbal pitch"),
	newDetail(CameraGimbalYaw, "sub_asset.payload_telemetry.camera_data.gimbal_yaw", "°", "Gimbal yaw"),
	newDetail(CameraGimbalRoll, "sub_asset.payload_telemetry.camera_data.gimbal_roll", "°", "Gimbal roll"),
	newDetail(CameraZoomFactor, "sub_asset.payload_telemetry.camera_data.zoom_factor", "", "Camera zoom factor"),
	newDetail(RangeFinderTargetLatitude, "sub_asset.payload_telemetry.range_finder_data.target_latitude", "°",
		"Latitude of the range finder's target"),
	newDetail(RangeFinderTargetLongitude, "sub_asset.payload_telemetry.range_finder_data.target_longitude", "°",
		"Longitude of the range finder's target"),
	newDetail(RangeFinderTargetDistance, "sub_asset.payload_telemetry.range_finder_data.target_distance", "m",
		"Distance to the range finder's target"),
	newDetail(RangeFinderTargetAltitude, "sub_asset.payload_telemetry.range_finder_data.target_altitude", "m",
		"Altitude of the range finder's target"),
	newDetail(SensorTargetTemperature, "sub_asset.payload_telemetry.sensor_data.target_temperature", "°C",
		"Temperature of the measured target"),
}

var fields = describeAll()

// Fields describes the keys ToSample writes into details, for CapabilitySet.telemetry_fields.
func Fields() []*capabilityv3.TelemetryField {
	out := make([]*capabilityv3.TelemetryField, len(fields))
	for i, f := range fields {
		out[i] = proto.Clone(f).(*capabilityv3.TelemetryField)
	}
	return out
}

// Field describes one key, or returns nil for a key this catalog does not know.
func Field(key string) *capabilityv3.TelemetryField {
	for _, f := range fields {
		if f.GetKey() == key {
			return proto.Clone(f).(*capabilityv3.TelemetryField)
		}
	}
	return nil
}

// RequestToSample is the sample of a v2 frame, or nil when the frame carries no telemetry.
func RequestToSample(req *livedatapb.ProduceTelemetryRequest) *telemetryv3.TelemetrySample {
	if req.GetData() == nil {
		return nil
	}
	return ToSample(req.GetData(), req.GetBase().GetSn())
}

// ToSample is the sample of v2 telemetry; fallbackSN names the asset when the telemetry does not.
func ToSample(t *livedatapb.Telemetry, fallbackSN string) *telemetryv3.TelemetrySample {
	sn := t.GetSn()
	if strings.TrimSpace(sn) == "" {
		sn = fallbackSN
	}
	sample := &telemetryv3.TelemetrySample{Asset: &commonv3.AssetRef{Sn: sn}, ObservedAt: t.GetTimestamp()}
	if t.Latitude != nil && t.Longitude != nil && isFinite(*t.Latitude) && isFinite(*t.Longitude) {
		sample.Position = &commonv3.GeoPoint{Latitude: *t.Latitude, Longitude: *t.Longitude}
		if t.AbsoluteAltitude != nil {
			sample.Position.Altitude = widened(*t.AbsoluteAltitude)
		}
	}
	if t.RelativeAltitude != nil {
		sample.RelativeAltitude = widened(*t.RelativeAltitude)
	}
	if t.Heading != nil {
		sample.HeadingDegrees = widened(*t.Heading)
	}
	if aircraft := t.GetSubAsset(); aircraft != nil {
		if aircraft.HorizontalSpeed != nil {
			sample.HorizontalSpeed = widened(*aircraft.HorizontalSpeed)
		}
		if aircraft.VerticalSpeed != nil {
			sample.VerticalSpeed = widened(*aircraft.VerticalSpeed)
		}
		if pct := aircraft.GetBatteryInformation().Percentage; pct != nil {
			if v, ok := parseNumber(*pct); ok {
				sample.BatteryPercent = &v
			}
		}
	} else if dock := t.GetAsset(); dock != nil && dock.SubAssetPercentage != nil {
		sample.BatteryPercent = widened(*dock.SubAssetPercentage)
	}
	values := map[string]*structpb.Value{}
	for _, d := range details {
		if v := read(t.ProtoReflect(), d.path); v != nil {
			values[d.key] = v
		}
	}
	if len(values) > 0 {
		sample.Details = &structpb.Struct{Fields: values}
	}
	return sample
}

// ToRequest is the v2 frame of a sample, as an adapter would have streamed it.
func ToRequest(sample *telemetryv3.TelemetrySample) *livedatapb.ProduceTelemetryRequest {
	base := &commonpb.RequestBase{
		Tid:       newID(),
		Sn:        sample.GetAsset().GetSn(),
		Timestamp: sample.GetObservedAt(),
	}
	if id := sample.GetAsset().GetId(); strings.TrimSpace(id) != "" {
		base.AssetId = &id
	}
	t := ToTelemetry(sample)
	t.Id = newID()
	return &livedatapb.ProduceTelemetryRequest{
		Base:      base,
		Telemetry: &livedatapb.ProduceTelemetryRequest_Data{Data: t},
	}
}

// ToTelemetry is the v2 telemetry of a sample.
func ToTelemetry(sample *telemetryv3.TelemetrySample) *livedatapb.Telemetry {
	t := &livedatapb.Telemetry{Sn: sample.GetAsset().GetSn(), Timestamp: sample.GetObservedAt()}
	if p := sample.GetPosition(); p != nil {
		lat, lon := p.GetLatitude(), p.GetLongitude()
		t.Latitude, t.Longitude = &lat, &lon
		if p.Altitude != nil {
			t.AbsoluteAltitude = narrowed(*p.Altitude)
		}
	}
	if sample.RelativeAltitude != nil {
		t.RelativeAltitude = narrowed(*sample.RelativeAltitude)
	}
	if sample.HeadingDegrees != nil {
		t.Heading = narrowed(*sample.HeadingDegrees)
	}

	values := sample.GetDetails().GetFields()
	src := sourceOf(sample, values)
	for _, d := range details {
		if v, ok := values[d.key]; ok && (d.source == sourceNone || d.source == src) {
			write(t.ProtoReflect(), d.path, v)
		}
	}
	switch src {
	case sourceSubAsset:
		aircraft := t.GetSubAsset()
		if aircraft == nil {
			aircraft = &livedatapb.SubAssetTelemetryDetails{}
			t.Source = &livedatapb.Telemetry_SubAsset{SubAsset: aircraft}
		}
		if sample.HorizontalSpeed != nil {
			aircraft.HorizontalSpeed = narrowed(*sample.HorizontalSpeed)
		}
		if sample.VerticalSpeed != nil {
			aircraft.VerticalSpeed = narrowed(*sample.VerticalSpeed)
		}
		if sample.BatteryPercent != nil {
			if aircraft.BatteryInformation == nil {
				aircraft.BatteryInformation = &livedatapb.SubAssetTelemetryDetails_SubAssetBatteryInformation{}
			}
			pct := formatNumber(*sample.BatteryPercent)
			aircraft.BatteryInformation.Percentage = &pct
		}
	case sourceAsset:
		if sample.BatteryPercent != nil {
			dock := t.GetAsset()
			if dock == nil {
				dock = &livedatapb.AssetTelemetryDetails{}
				t.Source = &livedatapb.Telemetry_Asset{Asset: dock}
			}
			dock.SubAssetPercentage = narrowed(*sample.BatteryPercent)
		}
	}
	return t
}

// sourceOf: an aircraft's sample moves; a dock's or a single device's carries its own keys or only a battery.
func sourceOf(sample *telemetryv3.TelemetrySample, values map[string]*structpb.Value) source {
	if sample.HorizontalSpeed != nil || sample.VerticalSpeed != nil || carries(values, sourceSubAsset) {
		return sourceSubAsset
	}
	if carries(values, sourceAsset) || sample.BatteryPercent != nil {
		return sourceAsset
	}
	return sourceNone
}

func carries(values map[string]*structpb.Value, src source) bool {
	for _, d := range details {
		if _, ok := values[d.key]; ok && d.source == src {
			return true
		}
	}
	return false
}

func read(msg protoreflect.Message, path []protoreflect.FieldDescriptor) *structpb.Value {
	current := msg
	for _, step := range path[:len(path)-1] {
		if !current.Has(step) {
			return nil
		}
		current = current.Get(step).Message()
	}
	leaf := path[len(path)-1]
	if !current.Has(leaf) {
		return nil
	}
	return toValue(leaf, current.Get(leaf))
}

func write(root protoreflect.Message, path []protoreflect.FieldDescriptor, v *structpb.Value) {
	leaf := path[len(path)-1]
	converted, ok := fromValue(leaf, v)
	if !ok {
		return
	}
	current := root
	for _, step := range path[:len(path)-1] {
		current = current.Mutable(step).Message()
	}
	current.Set(leaf, converted)
}

func toValue(fd protoreflect.FieldDescriptor, v protoreflect.Value) *structpb.Value {
	switch fd.Kind() {
	case protoreflect.FloatKind:
		return number(*widened(float32(v.Float())))
	case protoreflect.DoubleKind:
		return number(v.Float())
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return number(float64(v.Int()))
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind, protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return number(float64(v.Uint()))
	case protoreflect.BoolKind:
		return structpb.NewBoolValue(v.Bool())
	case protoreflect.StringKind:
		return structpb.NewStringValue(v.String())
	case protoreflect.EnumKind:
		ev := fd.Enum().Values().ByNumber(v.Enum())
		if ev == nil {
			return nil
		}
		return structpb.NewStringValue(enumName(fd.Enum(), ev))
	case protoreflect.MessageKind:
		if fd.Message().FullName() == timestampName {
			ts := &timestamppb.Timestamp{}
			proto.Merge(ts, v.Message().Interface())
			return structpb.NewStringValue(formatTimestamp(ts))
		}
	}
	return nil
}

func fromValue(fd protoreflect.FieldDescriptor, v *structpb.Value) (protoreflect.Value, bool) {
	n, isNumber := v.GetKind().(*structpb.Value_NumberValue)
	s, isString := v.GetKind().(*structpb.Value_StringValue)
	switch fd.Kind() {
	case protoreflect.FloatKind:
		if isNumber {
			return protoreflect.ValueOfFloat32(float32(n.NumberValue)), true
		}
	case protoreflect.DoubleKind:
		if isNumber {
			return protoreflect.ValueOfFloat64(n.NumberValue), true
		}
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		if isNumber {
			return protoreflect.ValueOfInt32(int32(math.Round(n.NumberValue))), true
		}
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		if isNumber {
			return protoreflect.ValueOfInt64(int64(math.Round(n.NumberValue))), true
		}
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		if isNumber {
			return protoreflect.ValueOfUint32(uint32(math.Round(n.NumberValue))), true
		}
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		if isNumber {
			return protoreflect.ValueOfUint64(uint64(math.Round(n.NumberValue))), true
		}
	case protoreflect.BoolKind:
		if b, ok := v.GetKind().(*structpb.Value_BoolValue); ok {
			return protoreflect.ValueOfBool(b.BoolValue), true
		}
	case protoreflect.StringKind:
		if isString {
			return protoreflect.ValueOfString(s.StringValue), true
		}
	case protoreflect.EnumKind:
		if isString {
			if ev := enumValue(fd.Enum(), s.StringValue); ev != nil {
				return protoreflect.ValueOfEnum(ev.Number()), true
			}
		}
	case protoreflect.MessageKind:
		if isString && fd.Message().FullName() == timestampName {
			if t, err := time.Parse(time.RFC3339Nano, s.StringValue); err == nil {
				return protoreflect.ValueOfMessage(timestamppb.New(t).ProtoReflect()), true
			}
		}
	}
	return protoreflect.Value{}, false
}

func newID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

var timestampName = (&timestamppb.Timestamp{}).ProtoReflect().Descriptor().FullName()

func number(v float64) *structpb.Value {
	if !isFinite(v) {
		return nil
	}
	return structpb.NewNumberValue(v)
}

func isFinite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

// widened is a float32 as the decimal it was written as (0.1 is 0.1, not 0.10000000149011612).
func widened(v float32) *float64 {
	w, _ := strconv.ParseFloat(strconv.FormatFloat(float64(v), 'g', -1, 32), 64)
	return &w
}

func narrowed(v float64) *float32 {
	n := float32(v)
	return &n
}

func parseNumber(text string) (float64, bool) {
	v, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
	return v, err == nil && isFinite(v)
}

func formatNumber(v float64) string {
	if v == math.Trunc(v) && math.Abs(v) < 1e15 {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// formatTimestamp writes RFC 3339 in UTC with 0, 3, 6 or 9 fraction digits, as protobuf's JSON mapping does.
func formatTimestamp(ts *timestamppb.Timestamp) string {
	t := ts.AsTime().UTC()
	layout := "2006-01-02T15:04:05"
	switch nanos := ts.GetNanos(); {
	case nanos == 0:
	case nanos%1_000_000 == 0:
		layout += ".000"
	case nanos%1_000 == 0:
		layout += ".000000"
	default:
		layout += ".000000000"
	}
	return t.Format(layout + "Z")
}

func enumName(enum protoreflect.EnumDescriptor, v protoreflect.EnumValueDescriptor) string {
	return strings.TrimPrefix(string(v.Name()), enumPrefix(enum))
}

func enumValue(enum protoreflect.EnumDescriptor, name string) protoreflect.EnumValueDescriptor {
	if v := enum.Values().ByName(protoreflect.Name(enumPrefix(enum) + name)); v != nil {
		return v
	}
	return enum.Values().ByName(protoreflect.Name(name))
}

// enumPrefix is what every value name of an enum starts with, up to an underscore: COVER_STATE_.
func enumPrefix(enum protoreflect.EnumDescriptor) string {
	values := enum.Values()
	prefix := string(values.Get(0).Name())
	for i := 0; i < values.Len(); i++ {
		for !strings.HasPrefix(string(values.Get(i).Name()), prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}
	return prefix[:strings.LastIndex(prefix, "_")+1]
}

func describeAll() []*capabilityv3.TelemetryField {
	out := make([]*capabilityv3.TelemetryField, 0, len(details))
	for _, d := range details {
		out = append(out, describe(d))
	}
	return out
}

func describe(d detail) *capabilityv3.TelemetryField {
	leaf := d.path[len(d.path)-1]
	f := &capabilityv3.TelemetryField{Key: d.key, Unit: d.unit, Description: d.description}
	switch leaf.Kind() {
	case protoreflect.FloatKind, protoreflect.DoubleKind,
		protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind,
		protoreflect.Uint32Kind, protoreflect.Fixed32Kind, protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		f.Type = capabilityv3.TelemetryValueType_TELEMETRY_VALUE_TYPE_NUMBER
	case protoreflect.BoolKind:
		f.Type = capabilityv3.TelemetryValueType_TELEMETRY_VALUE_TYPE_BOOLEAN
	case protoreflect.EnumKind:
		f.Type = capabilityv3.TelemetryValueType_TELEMETRY_VALUE_TYPE_STRING
		values := leaf.Enum().Values()
		for i := 0; i < values.Len(); i++ {
			f.AllowedValues = append(f.AllowedValues, enumName(leaf.Enum(), values.Get(i)))
		}
	default:
		f.Type = capabilityv3.TelemetryValueType_TELEMETRY_VALUE_TYPE_STRING
	}
	return f
}

func newDetail(key, v2Path, unit, description string) detail {
	var path []protoreflect.FieldDescriptor
	msg := (&livedatapb.Telemetry{}).ProtoReflect().Descriptor()
	for _, name := range strings.Split(v2Path, ".") {
		if msg == nil {
			panic("telemetry: " + v2Path + " walks past a scalar")
		}
		fd := msg.Fields().ByName(protoreflect.Name(name))
		if fd == nil || fd.IsList() || fd.IsMap() {
			panic("telemetry: no singular field " + name + " on the path " + v2Path)
		}
		path = append(path, fd)
		msg = fd.Message()
	}
	src := sourceNone
	switch path[0].Name() {
	case "asset":
		src = sourceAsset
	case "sub_asset":
		src = sourceSubAsset
	}
	return detail{key: key, path: path, source: src, unit: unit, description: description}
}
