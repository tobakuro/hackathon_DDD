extends Button

# 移動先のシーンパス（ファイルパスを指定してください）
const NEXT_SCENE_PATH = "res://scenes/main_3d.tscn"

func _on_pressed() -> void:
	# 指定したシーンへ遷移する
	get_tree().change_scene_to_file(NEXT_SCENE_PATH)
