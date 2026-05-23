extends XROrigin3D

@export var move_speed := 2.0

@onready var camera: XRCamera3D = $XRCamera3D
@onready var left_hand: XRController3D = $LeftHand
@onready var right_hand: XRController3D = $RightHand
@onready var left_hand_area: Area3D = $LeftHand/Area3D
@onready var right_hand_area: Area3D = $RightHand/Area3D

var _left_trigger_prev := false
var _right_trigger_prev := false


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

	print("%s picked by: %s" % [gopher.name, hand.name])
