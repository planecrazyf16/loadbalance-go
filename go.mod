module loadbalance

go 1.23.0

replace serverpool => ./serverpool

require (
	consistenthash v0.0.0-00010101000000-000000000000
	serverpool v0.0.0-00010101000000-000000000000
)

replace consistenthash => ./consistenthash
