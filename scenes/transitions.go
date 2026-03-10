package scenes

// NewMainMenuFactory returns a closure that creates a new MenuScene.
// Use this anywhere a "return to main menu" transition is needed,
// so all callers share the same factory logic.
func NewMainMenuFactory(sc SceneChanger) func() interface{} {
	return func() interface{} {
		return NewMenuScene(sc)
	}
}
