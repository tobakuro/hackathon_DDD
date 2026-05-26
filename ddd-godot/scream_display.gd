extends Node3D

@export var websocket_url := "ws://localhost:8082/ws"
@export var dock_path: NodePath = NodePath("../Docker_kun")
@export var cleanup_delay := 1.0
@export var reconnect_delay := 2.0
@export var ring_radius := 1.9
@export var height_offset := 1.2

@export var random_in_view := true
@export var random_screen_padding := 0.12
@export var random_distance_min := 2.0
@export var random_distance_max := 6.0
@export var vertical_shift := -0.3

@onready var dock: Node3D = get_node(dock_path)

var _socket := WebSocketPeer.new()
var _bubbles: Dictionary = {}
var _reconnect_cooldown := 0.0
var _was_socket_open := false

# Local alignment constants (use integers to avoid missing identifier issues in some analyzers)
const HA_CENTER := 1
const VA_CENTER := 1


func _ready():
	randomize()
	websocket_url = _resolve_websocket_url()
	_try_connect()


func _process(delta):
	_socket.poll()
	_sync_socket_state_log()
	_handle_packets()
	_handle_reconnect(delta)


func _try_connect():
	var state := _socket.get_ready_state()
	if state == WebSocketPeer.STATE_CONNECTING or state == WebSocketPeer.STATE_OPEN:
		return

	_socket = WebSocketPeer.new()

	var error := _socket.connect_to_url(websocket_url)
	if error != OK:
		push_warning("WebSocket接続に失敗しました: %s (error %d)" % [websocket_url, error])
		_reconnect_cooldown = reconnect_delay
		return


func _resolve_websocket_url() -> String:
	var env_url := str(OS.get_environment("SCREAM_WEBSOCKET_URL")).strip_edges()
	if not env_url.is_empty():
		return env_url

	var project_url := str(ProjectSettings.get_setting("application/config/scream_websocket_url", "")).strip_edges()
	if not project_url.is_empty():
		return project_url

	return websocket_url


func _handle_reconnect(delta):
	var state := _socket.get_ready_state()
	if state == WebSocketPeer.STATE_OPEN:
		_reconnect_cooldown = 0.0
		_was_socket_open = true
		return

	if state == WebSocketPeer.STATE_CONNECTING:
		return

	if _reconnect_cooldown > 0.0:
		_reconnect_cooldown -= delta
		return

	_try_connect()


func _sync_socket_state_log():
	var is_open := _socket.get_ready_state() == WebSocketPeer.STATE_OPEN
	if is_open and not _was_socket_open:
		print("WebSocket接続確立: ", websocket_url)
	_was_socket_open = is_open


func _handle_packets():
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		return

	while _socket.get_available_packet_count() > 0:
		var packet := _socket.get_packet()
		if packet.is_empty():
			continue

		var text := packet.get_string_from_utf8()
		var parsed: Variant = JSON.parse_string(text)
		if typeof(parsed) != TYPE_DICTIONARY:
			continue

		_handle_event(parsed)


func _handle_event(event: Dictionary):
	var id := str(event.get("id", ""))
	if id.is_empty():
		return

	var event_type := str(event.get("type", ""))
	match event_type:
		"start":
			var bubble := _ensure_bubble(id)
			bubble.text = ""
			bubble.visible = true
			bubble.position = _bubble_position_for(id)
		"char":
			var current := _ensure_bubble(id)
			current.text += str(event.get("char", ""))
		"end":
			_schedule_cleanup(id)
		"error":
			var failed := _ensure_bubble(id)
			failed.text = str(event.get("message", ""))
			_schedule_cleanup(id)
		_:
			pass


func _ensure_bubble(id: String) -> Label3D:
	if _bubbles.has(id):
		return _bubbles[id]

	# Create a container so we can add a subtle shadow behind the main label.
	var container := Node3D.new()
	container.name = "ScreamContainer_%s" % id

	# Shadow label (slightly larger/darker, placed behind)
	var shadow := Label3D.new()
	shadow.name = "Shadow"
	shadow.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	shadow.fixed_size = true
	shadow.no_depth_test = true
	shadow.pixel_size = 0.011
	shadow.font_size = 36
	shadow.modulate = Color(0.0, 0.0, 0.0, 0.85)
	shadow.horizontal_alignment = HA_CENTER
	shadow.vertical_alignment = VA_CENTER
	# slight offset for drop-shadow effect
	shadow.position = Vector3(0.02, -0.02, 0.02)
	container.add_child(shadow)

	# Main readable label
	var main := Label3D.new()
	main.name = "Main"
	main.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	main.fixed_size = true
	main.no_depth_test = true
	main.pixel_size = 0.01
	main.font_size = 36
	main.modulate = Color(1.0, 0.98, 0.9, 1.0)
	main.horizontal_alignment = HA_CENTER
	main.vertical_alignment = VA_CENTER
	main.position = Vector3(0, 0, 0)
	container.add_child(main)

	# Position the container and add to dock
	container.position = _bubble_position_for(id)
	container.visible = true

	dock.add_child(container)

	# Store the main label for back-compat (other code expects a Label3D)
	_bubbles[id] = main

	return main


func _bubble_position_for(id: String) -> Vector3:
	# If enabled, try to place the bubble at a random position inside the camera's view frustum.
	if random_in_view:
		var cam := get_viewport().get_camera_3d()
		if cam != null:
				var rect: Rect2 = get_viewport().get_visible_rect()
				var sz: Vector2 = rect.size
				var pad: float = clamp(random_screen_padding, 0.0, 0.45)
				var sx: float = pad + randf() * (1.0 - pad * 2.0)
				var sy: float = pad + randf() * (1.0 - pad * 2.0)
				var screen_pos: Vector2 = Vector2(sx * sz.x, sy * sz.y)
				var origin: Vector3 = cam.project_ray_origin(screen_pos)
				var dir: Vector3 = cam.project_ray_normal(screen_pos)
				var dist: float = random_distance_min + randf() * (random_distance_max - random_distance_min)
				return origin + dir * dist + Vector3(0, vertical_shift, 0)

	# Fallback deterministic ring placement around the dock (previous behavior)
	var numeric_id := int(id)
	if numeric_id == 0:
		numeric_id = abs(id.hash())

	var angle := deg_to_rad(float(numeric_id % 360))
	var radius := ring_radius + float(numeric_id % 3) * 0.25
	var height := height_offset + float((numeric_id / 3) % 3) * 0.18 + vertical_shift
	return Vector3(cos(angle) * radius, height, sin(angle) * radius)


func _schedule_cleanup(id: String):
	if not _bubbles.has(id):
		return

	var bubble: Label3D = _bubbles[id]
	var timer := get_tree().create_timer(cleanup_delay)
	timer.timeout.connect(_cleanup_bubble.bind(id, bubble))


func _cleanup_bubble(id: String, bubble: Label3D):
	# The stored `bubble` is the main Label3D child; free its parent container
	if is_instance_valid(bubble):
		var parent := bubble.get_parent()
		if is_instance_valid(parent):
			parent.queue_free()
	_bubbles.erase(id)
