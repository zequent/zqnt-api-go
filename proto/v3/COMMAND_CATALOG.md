# Command catalog (v3)

Every action on an asset is a capability with a dotted command id and runs through `ExecuteCommand`
(`zqnt.edge.v3`, `zqnt.control.v3`). Ids are **data, not proto**: an adapter registers a command in its
edge SDK and it is executable from Skills, the console, the client SDKs and the AI agent without any
other change. This catalog lists the ids the platform knows, so adapters reuse the same id for the same
action and the platform can rate its risk.

Rules
- `<domain>.<action>` or `<domain>.<object>.<action>`, lower_snake_case segments.
- Vendor-specific commands live under `vendor.<vendor>.<name>`; they never collide with shared ids.
- Events a command emits are named `<command_id>.<event>` (e.g. `flight.takeoff.completed`), declared in the capability's `events`.
- An adapter's own `CommandSafety` wins over this table; an id that is in neither counts as **CRITICAL**.
- Discovery never gates dispatch: an id missing here is still sent; the adapter decides.
- Changing the meaning of an id is a breaking change: add a new id and set `deprecated_by` on the old one.

Risk levels: **Observe** reads only · **Adjust** changes settings, nothing moves · **Move** moves the asset or a part of it · **Critical** irreversible or dangerous.

## Observe

| Command id | Undo | What it does |
|---|---|---|
| `camera.take_photo` | — | Takes a photo. |
| `camera.capture_photo` | — | Takes a photo (alias kept from 2.x; prefer camera.take_photo). |
| `camera.start_recording` | `camera.stop_recording` | Starts recording video. |
| `camera.stop_recording` | — | Stops recording video. |
| `stream.start` | `stream.stop` | Starts the live video stream. |
| `stream.stop` | — | Stops the live video stream. |
| `stream.split_screen` | — | Switches the live view between one and two cameras. |
| `detection.start` | `detection.stop` | Starts object detection on the live video. |
| `detection.stop` | — | Stops object detection. |
| `detection.status` | — | Reports whether detection is running. |
| `detection.wait` | — | Waits for a detection. |
| `detection.count` | — | Counts detected objects. |
| `detection.control` | — | Starts or stops object detection (2.x form). |

## Adjust

| Command id | Undo | What it does |
|---|---|---|
| `gimbal.look_at` | `gimbal.reset` | Points the camera at a location. |
| `gimbal.rotate` | `gimbal.reset` | Turns the camera. |
| `gimbal.reset` | — | Recentres the camera or points it straight down. |
| `gimbal.tracking` | — | Makes the camera follow a target. |
| `camera.change_lens` | — | Switches the camera lens. |
| `camera.change_zoom` | — | Changes the zoom. |
| `camera.thermal_palette` | — | Changes the thermal colour palette. |
| `speaker.audio.play` | `speaker.audio.stop` | Plays audio through the loudspeaker. |
| `speaker.audio.stop` | — | Stops loudspeaker playback. |
| `speaker.volume.set` | — | Changes the loudspeaker volume. |
| `searchlight.mode.set` | — | Changes the searchlight mode. |
| `searchlight.brightness.set` | — | Changes the searchlight brightness. |
| `widget.set` | — | Changes a payload setting. |
| `parachute.led.set` | — | Switches the parachute light. |
| `dock.light` | — | Switches the landing-pad light. |
| `dock.open_cover` | `dock.close_cover` | Opens the dock cover (no flight task running). |
| `dock.close_cover` | `dock.open_cover` | Closes the dock cover (no flight task running). |
| `dock.putter` | — | Moves the centring putters on the landing pad. |
| `dock.start_charging` | `dock.stop_charging` | Starts charging the docked aircraft. |
| `dock.stop_charging` | `dock.start_charging` | Stops charging the docked aircraft. |
| `asset.change_ac_mode` | — | Changes the dock's air conditioning mode. |
| `asset.alarm` | — | Switches the dock's audible and visual alarm. |
| `asset.remote_debug` | — | Switches the dock's remote debug mode (no flight task running). |
| `asset.property.<name>.set` | — | Changes one device setting (family, rated by prefix). |
| `battery.maintenance` | — | Switches the dock's battery maintenance. |
| `battery.store_mode` | — | Changes how the dock stores the battery. |
| `radar.look_at` | — | Turns the radar towards a location. |
| `radar.point` | — | Turns the radar. |
| `radar.set_mode` | — | Switches the radar between search and track. |
| `jammer.look_at` | — | Turns the jammer towards a location. |
| `jammer.set_bands` | — | Changes the bands the jammer can deny. |
| `mission.pause` | `mission.resume` | Pauses the running route (airborne). |
| `mission.prepare` | — | Uploads a route to the dock. |
| `flight.manual.enter` | `flight.manual.exit` | Starts manual control (stick input via StreamManualControl). |
| `flight.manual.exit` | — | Ends manual control. |

## Move

| Command id | Undo | What it does |
|---|---|---|
| `flight.takeoff` | `flight.return_to_home` | Takes off and climbs to the given height above the takeoff point (default 40 m). |
| `navigation.go_to` | `flight.return_to_home` | Flies to a position; altitude is relative to the takeoff point (default 40 m). |
| `navigation.update_target` | — | Changes the target of a running go-to. |
| `flight.return_to_home` | — | Returns to the dock/home point. |
| `flight.manual.input` | — | One manual-control input (stream form: StreamManualControl). |
| `flight.emergency_stop` | — | Stops and hovers immediately. |
| `mission.waypoint.execute` | `mission.stop` | Flies a waypoint route from the dock. |
| `mission.start` | `mission.stop` | Starts the prepared route. |
| `mission.resume` | `mission.pause` | Resumes a paused route. |
| `mission.stop` | — | Stops the route. |

## Critical

| Command id | Undo | What it does |
|---|---|---|
| `flight.land` | `flight.emergency_stop` | Lands where it is; must be recovered by hand. |
| `asset.reboot` | — | Restarts the device; offline until booted. |
| `asset.boot_sub_asset` | — | Powers the aircraft on or off (docked). |
| `link.mode` | — | Changes how the dock talks to the aircraft; a wrong mode can cut the link. |
| `jammer.enabled` | — | Starts or stops radio jamming; denies every receiver in the beam. |

## Vendor commands (rated by the adapter)

| Command id | What it does |
|---|---|
| `vendor.rns.send_raw` | Reticulum: sends a raw envelope. |
| `vendor.sapient.mode_change` | SAPIENT: changes the sensor mode. |
| `vendor.sapient.request` | SAPIENT: sends a task request. |
| `vendor.betaflight.arm` | Betaflight: arms the motors. |
| `vendor.betaflight.disarm` | Betaflight: disarms the motors. |

## Not commands any more

The 1.3.x Mission/Task RPCs (`PrepareTask`, `StartTask`, `StopTask`, `PauseTask`, `ResumeTask`) and every typed
command RPC of v2 (`TakeOff`, `GoTo`, `OpenCover`, …) have no v3 equivalent other than the ids above.
`asset.register`/`asset.deregister` are not commands: registration is part of the asset service (claim codes).
