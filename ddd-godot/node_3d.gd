extends XROrigin3D

@export var backend_host := "localhost"
@export var backend_port := 9000
@export var scale_proxy_port := 8082

@export var move_speed := 2.0

@onready var camera: XRCamera3D = $XRCamera3D
@onready var left_hand: XRController3D = $LeftHand
@onready var right_hand: XRController3D = $RightHand
@onready var left_hand_area: Area3D = $LeftHand/Area3D
@onready var right_hand_area: Area3D = $RightHand/Area3D

var _left_trigger_prev := false
var _right_trigger_prev := false
var _grabbed_gopher_count := 0
var _scale_use_proxy_only := false


func _ready():
	print("1: _ready started")

	var xr_interface = XRServer.find_interface("OpenXR")

	if xr_interface:
		print("2: OpenXR interface found")

		if xr_interface.initialize():
			print("3: OpenXR initialized")
			get_viewport().use_xr = true
			print("4: XR viewport enabled")
		else:
			print("ERROR: OpenXR initialize failed")
	else:
		print("ERROR: OpenXR interface NOT found")


func _process(delta):
	var stick := left_hand.get_vector2("primary")
	_handle_pickup()

	if stick.length() > 0.1:
		print("left stick: ", stick)

	var forward := -camera.global_transform.basis.z
	var right := camera.global_transform.basis.x

	forward.y = 0
	right.y = 0

	forward = forward.normalized()
	right = right.normalized()

	var move_dir := right * stick.x + forward * stick.y

	global_position += move_dir * move_speed * delta


func _handle_pickup():
	var left_trigger_now := left_hand.is_button_pressed("trigger_click")
	if left_trigger_now and not _left_trigger_prev:
		var left_target := _find_overlapping_gopher(left_hand_area)
		if left_target != null:
			_attach_gopher_to_hand(left_hand, left_target)

	var right_trigger_now := right_hand.is_button_pressed("trigger_click")
	if right_trigger_now and not _right_trigger_prev:
		var right_target := _find_overlapping_gopher(right_hand_area)
		if right_target != null:
			_attach_gopher_to_hand(right_hand, right_target)

	_left_trigger_prev = left_trigger_now
	_right_trigger_prev = right_trigger_now


func _find_overlapping_gopher(hand_area: Area3D) -> StaticBody3D:
	for body in hand_area.get_overlapping_bodies():
		if body is StaticBody3D and str(body.name).begins_with("GoGopher") and body.get_parent() != left_hand and body.get_parent() != right_hand:
			return body
	return null


func _attach_gopher_to_hand(hand: XRController3D, gopher: StaticBody3D):
	if gopher == null:
		return

	if gopher.get_parent() == left_hand or gopher.get_parent() == right_hand:
		return

	var grabbed_global := gopher.global_transform
	gopher.reparent(hand)
	gopher.global_transform = grabbed_global
	_grabbed_gopher_count += 1
	_request_scale(_grabbed_gopher_count)

	print("%s picked by: %s" % [gopher.name, hand.name])


func _request_scale(count: int):
	if _scale_use_proxy_only:
		_request_scale_with_fallback(_get_scale_proxy_url(), count, false)
		return

	_request_scale_with_fallback(_get_api_base_url() + "/scale", count, true)


func _request_scale_with_fallback(url: String, count: int, allow_fallback: bool):
	var request := HTTPRequest.new()
	add_child(request)
	request.request_completed.connect(_on_scale_request_completed.bind(request, count, url, allow_fallback))

	var headers := ["Content-Type: application/json"]
	var payload := JSON.stringify({"count": count})
	var error := request.request(
		url,
		headers,
		HTTPClient.METHOD_PUT,
		payload
	)

	if error != OK:
		print("/scale request failed to send (error: %d, count: %d, url=%s)" % [error, count, url])
		request.queue_free()
		if allow_fallback:
			_scale_use_proxy_only = true
			_request_scale_via_proxy(count, "send_error_%d" % error)


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


func _get_scale_proxy_url() -> String:
	return "http://%s:%d/scale-proxy" % [_resolve_backend_host(), scale_proxy_port]


func _on_scale_request_completed(result, response_code, _headers, body, request: HTTPRequest, count: int, url: String, allow_fallback: bool):
	if response_code == 200:
		print("/scale success: count=%d response=%s" % [count, body.get_string_from_utf8()])
	else:
		print("/scale failed: result=%d status=%d count=%d url=%s response=%s" % [result, response_code, count, url, body.get_string_from_utf8()])
		if allow_fallback:
			_scale_use_proxy_only = true
			_request_scale_via_proxy(count, "status_%d_result_%d" % [response_code, result])

	request.queue_free()


func _request_scale_via_proxy(count: int, reason: String):
	var proxy_url := _get_scale_proxy_url()
	print("Trying /scale proxy fallback: count=%d reason=%s url=%s" % [count, reason, proxy_url])
	_request_scale_with_fallback(proxy_url, count, false)
