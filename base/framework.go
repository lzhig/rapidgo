package base

type framework interface {
	run()
	quit()
	get_service_manager()
}
