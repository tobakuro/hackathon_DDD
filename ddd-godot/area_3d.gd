extends Area3D

@export var backend_host: String = "localhost"
@export var backend_port: int = 9000

# 接触ごとに送る攻撃回数のデフォルト値
@export var attack_count: int = 1

# クールダウン時間（秒）: 短時間の連続接触でAPIを連打しないよう制御
@export var cooldown_sec: float = 1.0

@onready var http_request := HTTPRequest.new()

var _is_requesting: bool = false
var _cooldown_timer: float = 0.0


func _ready():
	add_child(http_request)
	body_entered.connect(_on_body_entered)
	http_request.request_completed.connect(_on_request_completed)


func _process(delta):
	if _cooldown_timer > 0.0:
		_cooldown_timer -= delta


func _on_body_entered(body: Node):
	print("当たった:", body.name)

	if _is_requesting:
		print("リクエスト処理中のためスキップ")
		return

	if _cooldown_timer > 0.0:
		print("クールダウン中のためスキップ (残り %.2f 秒)" % _cooldown_timer)
		return

	# オブジェクト名に応じてAPIを切り替え
	if _should_trigger_attack(body):
		# 全attackerコンテナに攻撃命令
		call_api(
			"/attack",
			{"target": "all", "count": attack_count}
		)
	elif body.name == "ScaleTarget":
		# attackerコンテナをスケール
		call_api(
			"/scale",
			{"count": attack_count}
		)
	else:
		print("未定義のオブジェクト '%s' への接触 — APIはスキップ" % body.name)


func _should_trigger_attack(body: Node) -> bool:
	if body.name == "Docker_kun":
		return true

	if not str(body.name).begins_with("GoGopher"):
		return false

	var parent := body.get_parent()
	if not (parent is XRController3D):
		return false

	for overlapped_body in get_overlapping_bodies():
		if overlapped_body.name == "Docker_kun":
			return true

	return false


func call_api(endpoint: String, data: Dictionary):
	var url = _get_api_base_url() + endpoint
	var headers = ["Content-Type: application/json"]
	var json_str = JSON.stringify(data)
	var method := HTTPClient.METHOD_POST
	if endpoint == "/scale":
		method = HTTPClient.METHOD_PUT

	print("APIリクエスト送信: method=", method, " url=", url, " ペイロード: ", json_str)

	var error = http_request.request(
		url,
		headers,
		method,
		json_str
	)

	if error != OK:
		print("APIリクエスト送信失敗 (エラーコード: ", error, ")")
		return

	_is_requesting = true
	_cooldown_timer = cooldown_sec


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


func _on_request_completed(result, response_code, _headers, body):
	_is_requesting = false

	var response_text = body.get_string_from_utf8()
	if response_code == 200:
		print("API成功 (200): ", response_text)
	else:
		print("API失敗 (HTTP ", response_code, "): ", response_text)
