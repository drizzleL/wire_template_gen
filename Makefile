wire:
	go run .
	wire template/m_wire.go
	mkdir gen
	mv template/wire_gen.go gen
