package event

const (
	TopicAppMounted  = "app.mounted"
	TopicAppShutdown = "app.shutdown"

	TopicAppFocused = "app.focused"
	TopicAppBlurred = "app.blurred"

	TopicProviderAdded   = "provider.added"
	TopicProviderRemoved = "provider.removed"

	TopicModelSelect  = "llm.model.select"
	TopicModelsLoaded = "llm.models.loaded"
)
