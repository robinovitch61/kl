# Plan: Replace kl Viewport with Bubbleo Components

## Overview

Replace kl's internal `viewport` and `filterable_viewport` packages with bubbleo's `viewport` and `filterableviewport` components.

## Key Interface Changes

| kl Interface | bubbleo Interface |
|--------------|-------------------|
| `RenderableComparable` | `viewport.Object` |
| `Render() linebuffer.LineBufferer` | `GetItem() item.Item` |
| `Equals(other interface{}) bool` | `SetSelectionComparator(CompareFn[T])` |
| `linebuffer.LineBuffer` | `item.SingleItem` / `item.Item` |

## Implementation Steps

### Phase 1: Add Dependency

1. **Add bubbleo to go.mod**
   - `go get github.com/robinovitch61/bubbleo`

### Phase 2: Update Core Types

2. **Update `k8s_log.Log`** (`internal/k8s/k8s_log/k8s_log.go`)
   - Replace `LineBuffer linebuffer.LineBuffer` with `Item item.SingleItem`
   - Update all places that create/access `LineBuffer`

3. **Update `entity.Entity`** (`internal/k8s/entity/entity.go`)
   - Change `Render() linebuffer.LineBufferer` to `GetItem() item.Item`
   - Return `item.NewItem(e.Repr())`
   - Keep `EqualTo()` for use with `SetSelectionComparator()`

4. **Update `model.PageLog`** (`internal/model/page_log.go`)
   - Change `Render() linebuffer.LineBufferer` to `GetItem() item.Item`
   - Replace `linebuffer.NewMulti()` with `item.NewMulti()`
   - Keep `Equals()` logic for comparator function

5. **Create new `RenderableString`** (new location TBD, possibly `internal/model/`)
   - Simple wrapper implementing `viewport.Object`
   ```go
   type RenderableString struct {
       Item item.Item
   }
   func (r RenderableString) GetItem() item.Item { return r.Item }
   ```

### Phase 3: Update Page Types

6. **Update `EntityPage`** (`internal/page/entities.go`)
   - Replace `filterable_viewport.FilterableViewport[entity.Entity]` with `*filterableviewport.Model[entity.Entity]`
   - Create viewport with `viewport.New[entity.Entity](width, height, opts...)`
   - Wrap with `filterableviewport.New(vp, opts...)`
   - Use `SetSelectionComparator()` for maintaining selection
   - Adapt filtering: bubbleo uses `item.ExtractExactMatches()`/`ExtractRegexMatches()` internally

7. **Update `LogsPage`** (`internal/page/logs.go`)
   - Similar changes as EntityPage
   - Use `viewport.WithStickyBottom()` for log streaming behavior

8. **Update `SingleLogPage`** (`internal/page/log.go`)
   - Use new RenderableString type
   - Create custom keymap with shift-modified navigation keys:
     ```go
     shiftKeyMap := viewport.KeyMap{
         Up:   key.NewBinding(key.WithKeys("shift+up", "shift+k")),
         Down: key.NewBinding(key.WithKeys("shift+down", "shift+j")),
         // ... etc
     }
     ```

### Phase 4: API Adaptations

9. **Keymap Mapping** - Update usages:
   - kl `SetStringToHighlight()` -> bubbleo `SetHighlights()` (manual highlight extraction)
   - kl `ScrollSoItemIdxInView()` -> bubbleo `EnsureItemInView()`
   - kl `SetContent()` -> bubbleo `SetObjects()`
   - kl `SetMaintainSelection()` -> bubbleo `SetSelectionComparator()`

10. **Style Adaptation** - Create helper to map kl styles to bubbleo:
    ```go
    func toViewportStyles(s style.Styles) viewport.Styles
    func toFilterableViewportStyles(s style.Styles) filterableviewport.Styles
    ```

11. **Filter Behavior Differences**:
    - kl's "show context" shows all items, highlights matches
    - bubbleo's "matching items only" hides non-matching items
    - May need to adjust `WithMatchingItemsOnly()` accordingly

### Phase 5: Cleanup

12. **Delete deprecated packages**:
    - `internal/viewport/` (entire directory)
    - `internal/filterable_viewport/` (entire directory)
    - `internal/filter/` (if no longer needed)
    - Evaluate if `internal/textinput/` is still needed

13. **Update imports** across all files

## Files to Modify

| File | Change |
|------|--------|
| `go.mod` | Add bubbleo dependency |
| `internal/k8s/k8s_log/k8s_log.go` | LineBuffer -> Item |
| `internal/k8s/entity/entity.go` | Implement Object interface |
| `internal/model/page_log.go` | Implement Object interface |
| `internal/page/entities.go` | Use bubbleo filterableviewport |
| `internal/page/logs.go` | Use bubbleo filterableviewport |
| `internal/page/log.go` | Use bubbleo filterableviewport |
| `internal/style/style.go` | Add style conversion helpers |

## Files to Delete

- `internal/viewport/` (entire directory)
- `internal/filterable_viewport/` (entire directory)
- `internal/filter/filter.go`

## Design Decisions

1. **Filtering**: Use bubbleo's built-in matching via `item.ExtractExactMatches()`/`ExtractRegexMatches()`. Entity and log representations will work with substring/regex matching on `ContentNoAnsi()`.

2. **Shift-navigation in SingleLogPage**: Create a custom keymap in kl's SingleLogPage that uses shift-modified keys (shift+up, shift+down, etc.) rather than extending bubbleo.

3. **Header embedding**: kl currently embeds filter in viewport header. Bubbleo shows filter on a separate line. This is an acceptable UI change.

## Verification

1. Run `go build ./...` to verify compilation
2. Manual testing of all three pages:
   - EntityPage: navigation, filtering, selection persistence
   - LogsPage: log streaming, sticky bottom, filtering
   - SingleLogPage: read-only browsing, shift-navigation
3. Run existing tests: `go test ./...`
