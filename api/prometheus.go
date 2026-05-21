package main

func getNowTarget() NowTarget {
	return NowTarget{
		Memory: 75.5,
		CPU:    60.3,
	}
}

func getAttackHistory() []AttackHistory {
	return []AttackHistory{
		{Timestamp: "2024-06-01T12:00:00Z", Memory: 70.2, CPU: 55.1},
		{Timestamp: "2024-06-01T12:01:00Z", Memory: 80.5, CPU: 65.3},
		{Timestamp: "2024-06-01T12:02:00Z", Memory: 90.1, CPU: 75.0},
	}
}
