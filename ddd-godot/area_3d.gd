extends Area3D

# 司令塔APIのベースURL（docker-compose でホスト側の 9000 番ポートを使用）
const API_BASE_URL = "http://localhost:9000"

# 接触ごとに送る攻撃回数のデフォルト値
@export var attack_count: int = 10

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
	match body.name:
		"Docker_kun":
			# 全attackerコンテナに攻撃命令
			call_api(
				"/attack",
				{"target": "all", "count": attack_count}
			)
		"ScaleTarget":
			# attackerコンテナをスケール
			call_api(
				"/scale",
				{"count": attack_count}
			)
		_:
			print("未定義のオブジェクト '%s' への接触 — APIはスキップ" % body.name)


func call_api(endpoint: String, data: Dictionary):
	var url = API_BASE_URL + endpoint
	var headers = ["Content-Type: application/json"]
	var json_str = JSON.stringify(data)

	print("APIリクエスト送信: POST ", url, " ペイロード: ", json_str)

	var error = http_request.request(
		url,
		headers,
		HTTPClient.METHOD_POST,
		json_str
	)

	if error != OK:
		print("APIリクエスト送信失敗 (エラーコード: ", error, ")")
		return

	_is_requesting = true
	_cooldown_timer = cooldown_sec


func _on_request_completed(result, response_code, _headers, body):
	_is_requesting = false

	var response_text = body.get_string_from_utf8()
	if response_code == 200:
		print("API成功 (200): ", response_text)
	else:
		print("API失敗 (HTTP ", response_code, "): ", response_text)
