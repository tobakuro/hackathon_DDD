extends XROrigin3D

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
