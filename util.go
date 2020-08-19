package openapi

func (oapi OpenAPI) log(msg string) {
	defer oapi.logger.Sync()

	sugar := oapi.logger.Sugar()
	sugar.Infof(msg)
}
