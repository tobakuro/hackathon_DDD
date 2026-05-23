extends XROrigin3D

@export var move_speed := 2.0

@onready var camera: XRCamera3D = $XRCamera3D
@onready var left_hand: XRController3D = $LeftHand
@onready var right_hand: XRController3D = $RightHand
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
