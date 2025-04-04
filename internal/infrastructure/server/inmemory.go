package server

import (
	"context"
	"sync"

	"github.com/FreePeak/cortex/internal/domain"
)

// InMemoryResourceRepository implements a ResourceRepository using in-memory storage.
type InMemoryResourceRepository struct {
	resources sync.Map
}

// NewInMemoryResourceRepository creates a new InMemoryResourceRepository.
func NewInMemoryResourceRepository() *InMemoryResourceRepository {
	return &InMemoryResourceRepository{}
}

// GetResource retrieves a resource by its URI.
func (r *InMemoryResourceRepository) GetResource(ctx context.Context, uri string) (*domain.Resource, error) {
	if resource, ok := r.resources.Load(uri); ok {
		res, ok := resource.(*domain.Resource)
		if !ok {
			return nil, domain.ErrInternal
		}
		return res, nil
	}
	return nil, domain.NewResourceNotFoundError(uri)
}

// ListResources returns all available resources.
func (r *InMemoryResourceRepository) ListResources(ctx context.Context) ([]*domain.Resource, error) {
	var resources []*domain.Resource
	r.resources.Range(func(_, value interface{}) bool {
		res, ok := value.(*domain.Resource)
		if !ok {
			// Skip invalid entries
			return true
		}
		resources = append(resources, res)
		return true
	})
	return resources, nil
}

// AddResource adds a new resource to the repository.
func (r *InMemoryResourceRepository) AddResource(ctx context.Context, resource *domain.Resource) error {
	r.resources.Store(resource.URI, resource)
	return nil
}

// DeleteResource removes a resource from the repository.
func (r *InMemoryResourceRepository) DeleteResource(ctx context.Context, uri string) error {
	if _, ok := r.resources.Load(uri); !ok {
		return domain.NewResourceNotFoundError(uri)
	}
	r.resources.Delete(uri)
	return nil
}

// InMemoryToolRepository implements a ToolRepository using in-memory storage.
type InMemoryToolRepository struct {
	tools sync.Map
}

// NewInMemoryToolRepository creates a new InMemoryToolRepository.
func NewInMemoryToolRepository() *InMemoryToolRepository {
	return &InMemoryToolRepository{}
}

// GetTool retrieves a tool by its name.
func (r *InMemoryToolRepository) GetTool(ctx context.Context, name string) (*domain.Tool, error) {
	// Try to get the tool with the exact name
	if tool, ok := r.tools.Load(name); ok {
		t, ok := tool.(*domain.Tool)
		if !ok {
			return nil, domain.ErrInternal
		}
		return t, nil
	}

	return nil, domain.NewToolNotFoundError(name)
}

// ListTools returns all available tools.
func (r *InMemoryToolRepository) ListTools(ctx context.Context) ([]*domain.Tool, error) {
	var tools []*domain.Tool
	r.tools.Range(func(_, value interface{}) bool {
		t, ok := value.(*domain.Tool)
		if !ok {
			// Skip invalid entries
			return true
		}
		tools = append(tools, t)
		return true
	})
	return tools, nil
}

// AddTool adds a new tool to the repository.
func (r *InMemoryToolRepository) AddTool(ctx context.Context, tool *domain.Tool) error {
	// Store the tool with its original name
	r.tools.Store(tool.Name, tool)

	return nil
}

// DeleteTool removes a tool from the repository.
func (r *InMemoryToolRepository) DeleteTool(ctx context.Context, name string) error {
	if _, ok := r.tools.Load(name); !ok {
		return domain.NewToolNotFoundError(name)
	}
	r.tools.Delete(name)
	return nil
}

// InMemoryPromptRepository implements a PromptRepository using in-memory storage.
type InMemoryPromptRepository struct {
	prompts sync.Map
}

// NewInMemoryPromptRepository creates a new InMemoryPromptRepository.
func NewInMemoryPromptRepository() *InMemoryPromptRepository {
	return &InMemoryPromptRepository{}
}

// GetPrompt retrieves a prompt by its name.
func (r *InMemoryPromptRepository) GetPrompt(ctx context.Context, name string) (*domain.Prompt, error) {
	if prompt, ok := r.prompts.Load(name); ok {
		p, ok := prompt.(*domain.Prompt)
		if !ok {
			return nil, domain.ErrInternal
		}
		return p, nil
	}
	return nil, domain.NewPromptNotFoundError(name)
}

// ListPrompts returns all available prompts.
func (r *InMemoryPromptRepository) ListPrompts(ctx context.Context) ([]*domain.Prompt, error) {
	var prompts []*domain.Prompt
	r.prompts.Range(func(_, value interface{}) bool {
		p, ok := value.(*domain.Prompt)
		if !ok {
			// Skip invalid entries
			return true
		}
		prompts = append(prompts, p)
		return true
	})
	return prompts, nil
}

// AddPrompt adds a new prompt to the repository.
func (r *InMemoryPromptRepository) AddPrompt(ctx context.Context, prompt *domain.Prompt) error {
	r.prompts.Store(prompt.Name, prompt)
	return nil
}

// DeletePrompt removes a prompt from the repository.
func (r *InMemoryPromptRepository) DeletePrompt(ctx context.Context, name string) error {
	if _, ok := r.prompts.Load(name); !ok {
		return domain.NewPromptNotFoundError(name)
	}
	r.prompts.Delete(name)
	return nil
}

// InMemorySessionRepository implements a SessionRepository using in-memory storage.
type InMemorySessionRepository struct {
	sessions sync.Map
}

// NewInMemorySessionRepository creates a new InMemorySessionRepository.
func NewInMemorySessionRepository() *InMemorySessionRepository {
	return &InMemorySessionRepository{}
}

// GetSession retrieves a session by its ID.
func (r *InMemorySessionRepository) GetSession(ctx context.Context, id string) (*domain.ClientSession, error) {
	if session, ok := r.sessions.Load(id); ok {
		s, ok := session.(*domain.ClientSession)
		if !ok {
			return nil, domain.ErrInternal
		}
		return s, nil
	}
	return nil, domain.NewSessionNotFoundError(id)
}

// ListSessions returns all active sessions.
func (r *InMemorySessionRepository) ListSessions(ctx context.Context) ([]*domain.ClientSession, error) {
	var sessions []*domain.ClientSession
	r.sessions.Range(func(_, value interface{}) bool {
		s, ok := value.(*domain.ClientSession)
		if !ok {
			// Skip invalid entries
			return true
		}
		sessions = append(sessions, s)
		return true
	})
	return sessions, nil
}

// AddSession adds a new session to the repository.
func (r *InMemorySessionRepository) AddSession(ctx context.Context, session *domain.ClientSession) error {
	r.sessions.Store(session.ID, session)
	return nil
}

// DeleteSession removes a session from the repository.
func (r *InMemorySessionRepository) DeleteSession(ctx context.Context, id string) error {
	if _, ok := r.sessions.Load(id); !ok {
		return domain.NewSessionNotFoundError(id)
	}
	r.sessions.Delete(id)
	return nil
}
