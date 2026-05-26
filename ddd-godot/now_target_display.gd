extends Label3D

@export var poll_interval := 1.0
@export var request_timeout := 2.0
@export var death_text := "死んでいます"
@export var backend_host := "localhost"
@export var backend_port := 9000

var _request := HTTPRequest.new()
var _timer := Timer.new()
var _is_requesting := false

const HA_CENTER := 1
const VA_CENTER := 1


func _ready():
	horizontal_alignment = HA_CENTER
	vertical_alignment = VA_CENTER
	fixed_size = true
	no_depth_test = true
	pixel_size = 0.008
	font_size = 84
	modulate = Color(1.0, 0.15, 0.15, 1.0)
	visible = false
	text = ""

	_request.timeout = request_timeout
	_request.request_completed.connect(_on_request_completed)
	add_child(_request)

	_timer.wait_time = poll_interval
	_timer.one_shot = false
	_timer.autostart = true
	_timer.timeout.connect(_poll_now_target)
	add_child(_timer)
	_timer.start()

	_poll_now_target()


func _poll_now_target():
	if _is_requesting:
		return

	var url := _get_api_base_url() + "/now-target"
	var error := _request.request(url)
	if error != OK:
		return

	_is_requesting = true


func _on_request_completed(result, response_code, _headers, body):
	_is_requesting = false

	if result != HTTPRequest.RESULT_SUCCESS or response_code != 200:
		return

	var parsed: Variant = JSON.parse_string(body.get_string_from_utf8())
	if typeof(parsed) != TYPE_DICTIONARY:
		return

	var is_alive := bool(parsed.get("is_alive", true))
	_set_dead_state(not is_alive)


func _set_dead_state(is_dead: bool):
	visible = is_dead
	if is_dead:
		text = death_text
	else:
		text = ""


func _get_api_base_url() -> String:
	return "http://%s:%d" % [_resolve_backend_host(), _resolve_backend_port()]


func _resolve_backend_host() -> String:
	var env_host := str(OS.get_environment("BACKEND_HOST")).strip_edges()
	if not env_host.is_empty():
		return env_host

	var project_host := str(ProjectSettings.get_setting("application/config/backend_host", "")).strip_edges()
	if not project_host.is_empty():
		return project_host

	return backend_host


func _resolve_backend_port() -> int:
	var env_port := str(OS.get_environment("BACKEND_PORT")).strip_edges()
	if not env_port.is_empty():
		return int(env_port)

	var project_port := str(ProjectSettings.get_setting("application/config/backend_port", "")).strip_edges()
	if not project_port.is_empty():
		return int(project_port)

	if OS.has_feature("android"):
		return 9001

	return backend_port