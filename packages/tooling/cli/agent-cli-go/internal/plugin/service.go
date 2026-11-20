package plugin

import (
	"fmt"
	"path"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/plugin"
	origin "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/origin/shared"
)

// Service installs, lists and removes plugins and marketplaces of one store.
type Service struct {
	Store   Store
	Files   Files
	Origins origin.Registry
}

// Fetched is a plugin fetched from its source and not yet installed. The caller
// shows it to the user, who then approves its digest.
type Fetched struct {
	Manifest     plugin.Manifest
	Capabilities []plugin.Capability
	Digest       string
	FileCount    int
	Source       string
	Marketplace  string
	Resolved     string
	tree         artifact.Tree
}

// splitListed separates name@marketplace. A spec whose parts are not both
// kebab-case names, such as user@host:path, is a source and not a listing.
func splitListed(spec string) (name, marketplace string, ok bool) {
	name, marketplace, found := strings.Cut(spec, "@")
	if !found || !artifact.ValidName(name) || !artifact.ValidName(marketplace) {
		return "", "", false
	}
	return name, marketplace, true
}

// Fetch reads the plugin named by spec: name@marketplace for a plugin a
// registered marketplace lists, otherwise a directory or a repository that is the
// plugin itself. Nothing is installed.
func (s Service) Fetch(spec string) (Fetched, error) {
	if name, marketplace, ok := splitListed(spec); ok {
		return s.fetchListed(name, marketplace)
	}
	checkout, err := s.checkout(spec)
	if err != nil {
		return Fetched{}, err
	}
	defer checkout.Release()
	tree, err := origin.ReadTree(s.Files, checkout.Dir)
	if err != nil {
		return Fetched{}, err
	}
	return describe(tree, checkout, "")
}

func (s Service) fetchListed(name, marketplace string) (Fetched, error) {
	registry, err := s.Store.Registry()
	if err != nil {
		return Fetched{}, err
	}
	source, ok := registry.Find(marketplace)
	if !ok {
		return Fetched{}, fmt.Errorf("marketplace %q is not registered; add it with `agent-cli marketplace add <source>`", marketplace)
	}
	checkout, listing, err := s.fetchMarketplace(source.Source)
	if err != nil {
		return Fetched{}, err
	}
	defer checkout.Release()
	entry, ok := listing.Find(name)
	if !ok {
		return Fetched{}, fmt.Errorf("marketplace %q lists no plugin %q", marketplace, name)
	}
	relative, err := entry.Path()
	if err != nil {
		return Fetched{}, err
	}
	if err := origin.CheckNoSymlinks(s.Files, checkout.Dir, relative); err != nil {
		return Fetched{}, err
	}
	tree, err := origin.ReadTree(s.Files, path.Join(checkout.Dir, relative))
	if err != nil {
		return Fetched{}, err
	}
	fetched, err := describe(tree, checkout, marketplace)
	if err != nil {
		return Fetched{}, err
	}
	if fetched.Manifest.Name != name {
		return Fetched{}, fmt.Errorf("marketplace %q lists %q, but the plugin names itself %q", marketplace, name, fetched.Manifest.Name)
	}
	return fetched, nil
}

// checkout makes the source named by spec available through its origin.
func (s Service) checkout(spec string) (origin.Checkout, error) {
	selected, err := s.Origins.Select(spec)
	if err != nil {
		return origin.Checkout{}, err
	}
	return selected.Fetch(spec)
}

// fetchMarketplace fetches a marketplace and reads its listing. The caller
// releases the checkout.
func (s Service) fetchMarketplace(spec string) (origin.Checkout, plugin.Marketplace, error) {
	checkout, err := s.checkout(spec)
	if err != nil {
		return origin.Checkout{}, plugin.Marketplace{}, err
	}
	files, err := origin.ReadFiles(s.Files, checkout.Dir, plugin.MarketplacePaths())
	if err == nil {
		var listing plugin.Marketplace
		if listing, err = plugin.ReadMarketplace(files); err == nil {
			return checkout, listing, nil
		}
	}
	checkout.Release()
	return origin.Checkout{}, plugin.Marketplace{}, err
}

func describe(tree artifact.Tree, checkout origin.Checkout, marketplace string) (Fetched, error) {
	manifest, err := plugin.ReadManifest(tree)
	if err != nil {
		return Fetched{}, err
	}
	return Fetched{
		Manifest:     manifest,
		Capabilities: plugin.Detect(tree, manifest),
		Digest:       tree.Digest(),
		FileCount:    len(tree),
		Source:       checkout.Source,
		Marketplace:  marketplace,
		Resolved:     checkout.Resolved,
		tree:         tree,
	}, nil
}

// Commit installs a fetched plugin. It fails closed unless the user approved the
// exact digest (checked first, so a preview names the digest at once), granted every capability the plugin carries, and installed every
// dependency before. A plugin of the same name is replaced, and the files of the
// old version are removed when no other plugin uses them.
func (s Service) Commit(fetched Fetched, approved string, allow []plugin.Capability) (plugin.Locked, error) {
	if err := plugin.CheckApproval(fetched.Digest, approved); err != nil {
		return plugin.Locked{}, err
	}
	if err := plugin.Authorize(fetched.Capabilities, allow); err != nil {
		return plugin.Locked{}, err
	}
	lock, err := s.Store.Lock()
	if err != nil {
		return plugin.Locked{}, err
	}
	for _, dependency := range fetched.Manifest.Dependencies {
		if _, ok := lock.Find(dependency); !ok {
			return plugin.Locked{}, fmt.Errorf("plugin %q depends on %q, which is not installed; install it first", fetched.Manifest.Name, dependency)
		}
	}
	record := plugin.Locked{
		Name:         fetched.Manifest.Name,
		Version:      fetched.Manifest.Version,
		Source:       fetched.Source,
		Marketplace:  fetched.Marketplace,
		Resolved:     fetched.Resolved,
		Digest:       fetched.Digest,
		Grants:       capabilityNames(fetched.Capabilities),
		Dependencies: fetched.Manifest.Dependencies,
	}
	previous, replaced := lock.Find(record.Name)
	if err := s.Store.Put(fetched.tree); err != nil {
		return plugin.Locked{}, err
	}
	next := lock.With(record)
	if err := s.Store.SaveLock(next); err != nil {
		return plugin.Locked{}, err
	}
	if replaced && previous.Digest != record.Digest {
		return record, s.collect(next, previous.Digest)
	}
	return record, nil
}

// collect removes the stored files of digest when no plugin of lock uses them.
func (s Service) collect(lock plugin.Lock, digest string) error {
	for _, record := range lock.Plugins {
		if record.Digest == digest {
			return nil
		}
	}
	return s.Store.Delete(digest)
}

// List returns the installed plugins.
func (s Service) List() ([]plugin.Locked, error) {
	lock, err := s.Store.Lock()
	if err != nil {
		return nil, err
	}
	return lock.Plugins, nil
}

// Remove uninstalls a plugin. It fails closed while another plugin depends on it.
func (s Service) Remove(name string) error {
	lock, err := s.Store.Lock()
	if err != nil {
		return err
	}
	record, ok := lock.Find(name)
	if !ok {
		return fmt.Errorf("plugin %q is not installed", name)
	}
	if dependents := lock.Dependents(name); len(dependents) > 0 {
		return fmt.Errorf("plugin %q is a dependency of %s; remove those first", name, strings.Join(dependents, ", "))
	}
	next, _ := lock.Without(name)
	if err := s.Store.SaveLock(next); err != nil {
		return err
	}
	return s.collect(next, record.Digest)
}

// Load returns the artifacts of the named plugins and of their dependencies. It
// fails closed on a plugin that is not installed, whose stored files changed,
// whose manifest names another plugin, or that carries a capability the user did
// not grant at install time.
func (s Service) Load(names []string) (artifact.Set, error) {
	closure, err := s.Locked(names)
	if err != nil {
		return artifact.Set{}, err
	}
	sets := make([]artifact.Set, 0, len(closure))
	for _, record := range closure {
		set, err := s.load(record)
		if err != nil {
			return artifact.Set{}, fmt.Errorf("plugin %q: %w", record.Name, err)
		}
		sets = append(sets, set)
	}
	return artifact.Merge(sets...)
}

// Locked returns the lock records of the named plugins and of every plugin they
// depend on, each with the digest of its stored content.
func (s Service) Locked(names []string) ([]plugin.Locked, error) {
	lock, err := s.Store.Lock()
	if err != nil {
		return nil, err
	}
	return lock.Closure(names)
}

func (s Service) load(record plugin.Locked) (artifact.Set, error) {
	tree, err := s.Store.Get(record.Digest)
	if err != nil {
		return artifact.Set{}, err
	}
	manifest, err := plugin.ReadManifest(tree)
	if err != nil {
		return artifact.Set{}, err
	}
	if manifest.Name != record.Name {
		return artifact.Set{}, fmt.Errorf("the stored files name the plugin %q", manifest.Name)
	}
	// DecodeLock already validated every grant.
	granted := make([]plugin.Capability, len(record.Grants))
	for index, grant := range record.Grants {
		granted[index] = plugin.Capability(grant)
	}
	if err := plugin.Authorize(plugin.Detect(tree, manifest), granted); err != nil {
		return artifact.Set{}, err
	}
	return artifact.FromTree(record.Name, tree)
}

// AddMarketplace registers the marketplace at spec under the name its listing
// gives it, and returns the listing.
func (s Service) AddMarketplace(spec string) (plugin.Source, plugin.Marketplace, error) {
	checkout, listing, err := s.fetchMarketplace(spec)
	if err != nil {
		return plugin.Source{}, plugin.Marketplace{}, err
	}
	defer checkout.Release()
	registry, err := s.Store.Registry()
	if err != nil {
		return plugin.Source{}, plugin.Marketplace{}, err
	}
	source := plugin.Source{Name: listing.Name, Source: checkout.Source}
	if err := s.Store.SaveRegistry(registry.With(source)); err != nil {
		return plugin.Source{}, plugin.Marketplace{}, err
	}
	return source, listing, nil
}

// Marketplaces returns the registered marketplaces.
func (s Service) Marketplaces() ([]plugin.Source, error) {
	registry, err := s.Store.Registry()
	if err != nil {
		return nil, err
	}
	return registry.Marketplaces, nil
}

// Browse fetches a registered marketplace and returns its listing.
func (s Service) Browse(name string) (plugin.Marketplace, error) {
	registry, err := s.Store.Registry()
	if err != nil {
		return plugin.Marketplace{}, err
	}
	source, ok := registry.Find(name)
	if !ok {
		return plugin.Marketplace{}, fmt.Errorf("marketplace %q is not registered", name)
	}
	checkout, listing, err := s.fetchMarketplace(source.Source)
	if err != nil {
		return plugin.Marketplace{}, err
	}
	checkout.Release()
	return listing, nil
}

// RemoveMarketplace unregisters a marketplace. It fails closed while an installed
// plugin came from it.
func (s Service) RemoveMarketplace(name string) error {
	registry, err := s.Store.Registry()
	if err != nil {
		return err
	}
	next, ok := registry.Without(name)
	if !ok {
		return fmt.Errorf("marketplace %q is not registered", name)
	}
	lock, err := s.Store.Lock()
	if err != nil {
		return err
	}
	var installed []string
	for _, record := range lock.Plugins {
		if record.Marketplace == name {
			installed = append(installed, record.Name)
		}
	}
	if len(installed) > 0 {
		return fmt.Errorf("marketplace %q is the source of %s; remove those plugins first", name, strings.Join(installed, ", "))
	}
	return s.Store.SaveRegistry(next)
}

func capabilityNames(capabilities []plugin.Capability) []string {
	names := make([]string, len(capabilities))
	for index, capability := range capabilities {
		names[index] = string(capability)
	}
	return names
}
