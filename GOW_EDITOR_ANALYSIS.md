# God of War 3D Editor - Architecture Analysis & Recommendations

**Date:** 2025-11-11
**Project:** god_of_war_browser
**Objective:** Create a comprehensive 3D editor for God of War game files

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Current System Analysis](#current-system-analysis)
3. [Integration Options Analysis](#integration-options-analysis)
4. [Recommended Architecture](#recommended-architecture)
5. [Implementation Roadmap](#implementation-roadmap)
6. [Technical Specifications](#technical-specifications)
7. [Risk Analysis](#risk-analysis)

---

## Executive Summary

The god_of_war_browser project provides a **solid foundation** for building a 3D editor for God of War games (GoW1, GoW2, GoW2018). The existing codebase includes:

- ✅ Complete parsing system for all major file formats (mesh, materials, textures, animations, scenes)
- ✅ WebGL-based 3D renderer with skeletal animation support
- ✅ GLTF/FBX/OBJ export capabilities
- ✅ Web-based architecture with Go backend + JavaScript frontend
- ✅ Modding support (texture upload, resource replacement)

**Key Gap:** The system currently lacks **write/serialization** capabilities for most formats and a proper **editor UI** with manipulation tools.

**Recommendation:** Build a **modern web-based 3D editor** leveraging the existing Go backend while replacing the frontend with a modern framework (React/Vue + Three.js) rather than attempting to integrate with external engines like Godot or Unity.

**Rationale:**
- Existing GLTF export enables external editor workflows
- Maintaining a unified codebase simplifies development
- Web-based approach provides cross-platform compatibility
- Direct access to game-specific formats without conversion overhead

---

## Current System Analysis

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    God of War Browser                        │
├─────────────────────────────────────────────────────────────┤
│  Frontend (JavaScript + WebGL)                               │
│  ├── Browser UI (jQuery-based)                               │
│  ├── 3D Renderer (Custom WebGL 1.0)                          │
│  ├── Animation System (Skeletal + Material)                  │
│  └── Resource Viewer (Tree + Properties)                     │
├─────────────────────────────────────────────────────────────┤
│  Backend (Go + Gorilla Mux)                                  │
│  ├── HTTP REST API (/json, /dump, /upload, /action)         │
│  ├── WAD Parser (Handler Registry Pattern)                   │
│  ├── Binary Parsers (150+ Go files)                          │
│  │   ├── MESH (0x0001000f) - Geometry                        │
│  │   ├── MAT (0x00000008) - Materials                        │
│  │   ├── TXR (0x00000007) - Textures                         │
│  │   ├── ANM (0x00000003) - Animations                       │
│  │   ├── OBJ (0x00040001) - Skeletons                        │
│  │   ├── INST (0x00020001) - Instances                       │
│  │   └── 50+ other formats                                   │
│  ├── Export System (GLTF, FBX, OBJ)                          │
│  └── VFS Abstraction (ISO/TOC/PSARC/Directory)               │
└─────────────────────────────────────────────────────────────┘
```

### Supported File Formats

#### Core 3D Assets
| Format | Magic | Description | Read | Write | Export |
|--------|-------|-------------|------|-------|--------|
| MESH | 0x0001000f | PS2 geometry (DMA packets) | ✅ | ❌ | ✅ GLTF/FBX/OBJ |
| GMDL | 0x0003000f | Alternative mesh format | ✅ | ❌ | ✅ GLTF/FBX/OBJ |
| MAT | 0x00000008 | Materials (multi-layer) | ✅ | ❌ | ✅ GLTF/FBX |
| TXR | 0x00000007 | Textures (references) | ✅ | ✅ | ✅ PNG |
| GFX | 0x00000009 | Texture data + palettes | ✅ | ✅ | ✅ PNG |
| OBJ | 0x00040001 | Skeleton/bone hierarchy | ✅ | ❌ | ✅ GLTF/FBX |
| ANM | 0x00000003 | Animations (12 types) | ✅ | ❌ | ❌ |
| MDL | 0x0002000f | Model container | ✅ | ❌ | ✅ GLTF/FBX |
| INST | 0x00020001 | Object instances | ✅ | ❌ | ✅ GLTF |
| CXT | 0x80000001 | Scene chunk/context | ✅ | ❌ | ✅ GLTF |

#### Other Assets
- **FLP (0x00000005):** UI/Font layouts - Read ✅ Write ✅
- **SBK/VAG:** Audio - Read ✅
- **SCR:** Scripts - Read ✅ (partial)
- **TWK:** Tweakable parameters - Read ✅ Write ✅ (YAML)
- **Collision:** Various formats - Read ✅

### Existing Capabilities

#### Strengths
1. **Complete Read Pipeline:** All major formats fully parsed
2. **GLTF Export:** Full scene export including skinning, hierarchy, materials
3. **Real-time Rendering:** WebGL viewer with animations
4. **Modding Support:** Texture replacement + resource upload
5. **Cross-version Support:** GoW1, GoW2, GoW2018 (partial)
6. **Clean Architecture:** Handler pattern makes adding formats easy

#### Limitations
1. **No Write System:** Cannot serialize back to binary (except TXR, FLP, TWK)
2. **Read-only UI:** No manipulation tools (transform, edit geometry)
3. **WebGL 1.0:** Limited to basic rendering (no PBR, shadows, compute)
4. **No Undo/Redo:** Missing fundamental editor feature
5. **No Animation Export:** GLTF export omits animation data
6. **PS2-Specific Complexity:** DMA packet compilation for mesh writing

### Key Technical Details

#### Data Structures

**Mesh Format:**
```
Mesh
├── Parts (one per bone/LOD group)
│   └── Groups (LOD levels)
│       └── Objects (drawable batches)
│           ├── Material ID
│           ├── Packets[] (PS2 DMA VIF/VU data)
│           │   ├── Vertices (X,Y,Z) - triangle strips
│           │   ├── UVs (U,V) per texture layer
│           │   ├── Normals (X,Y,Z)
│           │   ├── Blend Colors (R,G,B,A)
│           │   ├── Joint IDs (primary + secondary)
│           │   └── Weights (0.0-1.0)
│           └── Joint Mappers (bone remapping)
```

**Skeleton Format:**
```
Object
└── Joints[]
    ├── Name (string)
    ├── Parent (index)
    ├── Children (start/end indices)
    ├── Flags (animation types)
    ├── ParentToJoint Matrix (local transform)
    └── BindToJoint Matrix (inverse bind pose)
```

**Animation Format:**
```
Animations
├── DataTypes[] (rotation, position, UV, color, etc.)
└── Groups[]
    └── Acts[] (animation clips)
        ├── Duration (seconds)
        └── StateDescrs[] (per data type)
            ├── FrameTime (keyframe interval)
            └── Data (type-specific keyframes)
                ├── Type 0: Bone transforms (Q.14 fixed-point)
                ├── Type 3: Material colors
                ├── Type 8: UV scrolling
                └── Type 9: Texture flipbook
```

#### Binary Parsing Approach

The codebase uses three levels of abstraction:

1. **Low-level:** Direct byte manipulation
```go
magic := binary.LittleEndian.Uint32(buf[0:4])
```

2. **Mid-level:** Binary reader
```go
binary.Read(bytes.NewReader(buf), binary.LittleEndian, &struct)
```

3. **High-level:** BufStack with debug info
```go
bs := utils.NewBufStack("resource", data)
value := bs.SubBuf("name", offset).ReadU32()
```

**Key Insight:** To implement writing, we need to create **inverse serializers** for each format, converting Go structs back to binary data following the exact format specifications.

---

## Integration Options Analysis

### Option 1: Godot Engine Integration

#### Approach
Create a Godot plugin that imports/exports GoW formats via the existing GLTF pipeline.

#### Pros
- ✅ Professional editor out-of-the-box (viewport, hierarchy, properties)
- ✅ GDScript for rapid prototyping
- ✅ Built-in animation editor
- ✅ Cross-platform (Windows, Linux, macOS)
- ✅ MIT licensed (no restrictions)

#### Cons
- ❌ **GLTF conversion overhead:** Every edit requires round-trip conversion
- ❌ **Loss of fidelity:** GLTF doesn't support all GoW-specific features
  - PS2 DMA packet structure
  - Multiple blend color layers
  - Game-specific material flags
  - 12 animation types (only skeletal → GLTF)
- ❌ **Limited to GLTF subset:** Godot uses its own scene format internally
- ❌ **Dual codebase:** C++/GDScript plugin + Go backend
- ❌ **Network latency:** API calls to Go server add overhead
- ❌ **Complex build:** Users need Godot + plugin + Go server

**Verdict:** ❌ **Not Recommended** - Too much complexity for limited benefit. GLTF export already enables Godot workflow for users who want it.

---

### Option 2: Unity Editor Integration

#### Approach
Create Unity Editor scripts that communicate with Go backend via REST API.

#### Pros
- ✅ Industry-standard editor
- ✅ Powerful C# API
- ✅ Asset Store ecosystem
- ✅ Advanced rendering (URP/HDRP)

#### Cons
- ❌ **Proprietary license:** Unity Personal/Plus/Pro restrictions
- ❌ **Same GLTF limitations** as Godot option
- ❌ **Heavy runtime:** Unity Editor is resource-intensive
- ❌ **C# ↔ Go overhead:** Cross-process communication
- ❌ **Versioning issues:** Unity updates break Editor APIs frequently
- ❌ **Distribution complexity:** Users need Unity + custom package

**Verdict:** ❌ **Not Recommended** - All the same issues as Godot, plus licensing concerns.

---

### Option 3: Blender Plugin

#### Approach
Python addon that imports GoW files via GLTF, edits in Blender, exports back to Go server.

#### Pros
- ✅ Professional 3D modeling tools
- ✅ Python API (easy to extend)
- ✅ GPL license (free forever)
- ✅ Industry standard for game modding
- ✅ Large community

#### Cons
- ❌ **Not a game engine:** No real-time preview with game shaders
- ❌ **Same GLTF conversion issues**
- ❌ **Python ↔ Go integration:** Requires REST API or subprocess
- ❌ **Learning curve:** Blender's UI is complex for casual modders
- ❌ **No game-specific tools:** Would need to build custom UI panels

**Verdict:** ⚠️ **Optional Add-on** - Good for advanced mesh editing, but not as primary editor. Existing GLTF export already enables this workflow.

---

### Option 4: Web-Based Editor (Recommended)

#### Approach
Build a modern web application using the existing Go backend + new React/Vue + Three.js frontend.

#### Pros
- ✅ **Leverages existing codebase:** 90% of backend already done
- ✅ **Direct format access:** No lossy GLTF conversion
- ✅ **Single codebase:** Go + TypeScript
- ✅ **Cross-platform:** Runs in any modern browser
- ✅ **Zero installation:** No dependencies (or Electron for desktop)
- ✅ **Fast iteration:** Hot reload during development
- ✅ **Modern UI frameworks:** React/Vue + component libraries
- ✅ **Three.js ecosystem:** Mature WebGL library with extensive tools
  - TransformControls (gizmos)
  - OrbitControls (camera)
  - Loaders (GLTF, FBX, OBJ)
  - Post-processing
  - Animation system
- ✅ **Progressive enhancement:** Start simple, add features incrementally
- ✅ **WebGPU ready:** Can upgrade to WebGPU for modern rendering

#### Cons
- ⚠️ **Need to implement serialization:** Write system for all formats
- ⚠️ **Browser limitations:** No direct file system access (use File API)
- ⚠️ **Performance:** JavaScript slower than C++ (but sufficient for editor)

**Verdict:** ✅ **HIGHLY RECOMMENDED** - Best balance of effort vs. reward.

---

## Recommended Architecture

### System Design: Web-Based 3D Editor

```
┌───────────────────────────────────────────────────────────────────┐
│                     GOD OF WAR EDITOR                              │
├───────────────────────────────────────────────────────────────────┤
│                                                                    │
│  FRONTEND (React + TypeScript + Three.js)                         │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │  UI Layer (React Components)                                 │ │
│  │  ├── Asset Browser (tree view with search/filter)            │ │
│  │  ├── Scene Hierarchy (drag-drop, multi-select)               │ │
│  │  ├── Properties Panel (context-sensitive)                    │ │
│  │  ├── Timeline/Dope Sheet (animation editor)                  │ │
│  │  ├── Material Editor (node-based or layer view)              │ │
│  │  ├── Toolbar (transform modes, tools)                        │ │
│  │  └── Menu Bar (File, Edit, View, Tools, Help)               │ │
│  ├──────────────────────────────────────────────────────────────┤ │
│  │  State Management (Zustand or Redux)                         │ │
│  │  ├── Project state (loaded WADs, active scene)               │ │
│  │  ├── Selection state (selected objects, bones, etc.)         │ │
│  │  ├── Undo/Redo stack (command pattern)                       │ │
│  │  ├── View state (camera, visibility, render mode)            │ │
│  │  └── Edit history (auto-save, versioning)                    │ │
│  ├──────────────────────────────────────────────────────────────┤ │
│  │  3D Viewport (Three.js)                                       │ │
│  │  ├── Scene Graph (THREE.Scene)                               │ │
│  │  │   ├── Cameras (Perspective, Orthographic)                 │ │
│  │  │   ├── Lights (Ambient, Directional, Point)                │ │
│  │  │   ├── Meshes (SkinnedMesh, Mesh, LOD)                     │ │
│  │  │   └── Helpers (Grid, Axes, Skeleton, Gizmos)              │ │
│  │  ├── Renderer (THREE.WebGLRenderer / WebGPURenderer)         │ │
│  │  ├── Controls (OrbitControls, TransformControls)             │ │
│  │  ├── Animation Mixer (THREE.AnimationMixer)                  │ │
│  │  └── Raycaster (object picking, snapping)                    │ │
│  ├──────────────────────────────────────────────────────────────┤ │
│  │  Data Layer (API Client + Cache)                             │ │
│  │  ├── REST API client (fetch/axios)                           │ │
│  │  ├── WebSocket client (live updates, collab?)               │ │
│  │  ├── Resource cache (textures, meshes, materials)            │ │
│  │  └── Serialization (JSON ↔ Three.js objects)                │ │
│  └──────────────────────────────────────────────────────────────┘ │
│                             ↕ HTTP/WS                              │
├───────────────────────────────────────────────────────────────────┤
│                                                                    │
│  BACKEND (Go + Gorilla Mux)                                       │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │  HTTP API (Extended)                                          │ │
│  │  ├── READ Endpoints (Existing)                               │ │
│  │  │   ├── GET /json/pack/{file}                               │ │
│  │  │   ├── GET /json/pack/{file}/{id}                          │ │
│  │  │   ├── GET /action/{file}/{id}/{action}                    │ │
│  │  │   └── GET /dump/pack/{file}/{id}                          │ │
│  │  ├── WRITE Endpoints (NEW)                                   │ │
│  │  │   ├── POST /api/mesh/{file}/{id} (update mesh)            │ │
│  │  │   ├── POST /api/material/{file}/{id}                      │ │
│  │  │   ├── POST /api/skeleton/{file}/{id}                      │ │
│  │  │   ├── POST /api/animation/{file}/{id}                     │ │
│  │  │   ├── POST /api/instance/{file}/{id}                      │ │
│  │  │   └── POST /api/scene/{file}/{id}                         │ │
│  │  ├── PROJECT Endpoints (NEW)                                 │ │
│  │  │   ├── POST /api/project/new                               │ │
│  │  │   ├── GET /api/project/{id}/export (GLTF/FBX)             │ │
│  │  │   ├── POST /api/project/{id}/import (GLTF)                │ │
│  │  │   └── GET /api/project/{id}/validate                      │ │
│  │  └── UTILITY Endpoints                                        │ │
│  │      ├── POST /api/tools/optimize-mesh                        │ │
│  │      ├── POST /api/tools/generate-lod                         │ │
│  │      └── POST /api/tools/bake-animations                      │ │
│  ├──────────────────────────────────────────────────────────────┤ │
│  │  Parsing Layer (Existing)                                     │ │
│  │  ├── Handler Registry (magic → parser)                       │ │
│  │  ├── Binary Parsers (150+ formats)                           │ │
│  │  └── Export System (GLTF/FBX/OBJ)                            │ │
│  ├──────────────────────────────────────────────────────────────┤ │
│  │  Serialization Layer (NEW)                                    │ │
│  │  ├── Mesh Serializer (struct → binary)                       │ │
│  │  │   ├── DMA Packet Compiler (PS2)                           │ │
│  │  │   ├── GPU Buffer Writer (GoW2018)                         │ │
│  │  │   └── Validation (bone weights, UVs, indices)             │ │
│  │  ├── Material Serializer                                      │ │
│  │  ├── Animation Serializer                                     │ │
│  │  │   ├── Keyframe optimizer                                  │ │
│  │  │   └── Fixed-point converter (Q.14)                        │ │
│  │  ├── Skeleton Serializer                                      │ │
│  │  └── Scene Serializer (INST, CXT)                            │ │
│  ├──────────────────────────────────────────────────────────────┤ │
│  │  Tools Layer (NEW)                                            │ │
│  │  ├── Mesh Optimizer (triangle strips, vertex cache)          │ │
│  │  ├── LOD Generator (mesh decimation)                         │ │
│  │  ├── Texture Compressor (PS2 swizzle, palettes)              │ │
│  │  ├── Animation Baker (sample → keyframes)                    │ │
│  │  └── Collision Generator (convex hull, AABB)                 │ │
│  ├──────────────────────────────────────────────────────────────┤ │
│  │  VFS Layer (Existing)                                         │ │
│  │  ├── ISO Driver                                              │ │
│  │  ├── TOC/PAK Driver                                          │ │
│  │  ├── PSARC Driver                                            │ │
│  │  └── Directory Driver                                        │ │
│  └──────────────────────────────────────────────────────────────┘ │
│                                                                    │
├───────────────────────────────────────────────────────────────────┤
│  STORAGE (File System)                                             │
│  ├── Original game files (.iso, .pak, .toc, .psarc)               │
│  ├── Extracted WAD files (.wad)                                   │
│  ├── Project files (.gowproj - JSON metadata)                     │
│  ├── Asset cache (textures, thumbnails)                           │
│  └── Exports (GLTF, FBX, OBJ)                                     │
└───────────────────────────────────────────────────────────────────┘
```

---

### Technology Stack

#### Frontend
| Component | Technology | Justification |
|-----------|-----------|---------------|
| Framework | **React 18** + TypeScript | Industry standard, huge ecosystem, hooks-based |
| 3D Engine | **Three.js r160+** | Mature, well-documented, extensive features |
| UI Library | **Ant Design** or **Chakra UI** | Professional components, customizable |
| State | **Zustand** | Lightweight, easy undo/redo with immer |
| Build Tool | **Vite** | Fast HMR, modern, excellent DX |
| Styling | **Tailwind CSS** + CSS Modules | Utility-first, responsive, maintainable |
| Icons | **React Icons** or **Lucide** | Comprehensive icon sets |
| Forms | **React Hook Form** | Performant, minimal re-renders |
| Panels | **react-resizable-panels** | Dockable layout like IDE |

**Alternative:** Vue 3 + Composition API (if team prefers Vue)

#### Backend
| Component | Technology | Status |
|-----------|-----------|--------|
| Language | **Go 1.21+** | ✅ Existing |
| HTTP Router | **Gorilla Mux** | ✅ Existing |
| WebSocket | **gorilla/websocket** | 🆕 Add for live updates |
| Validation | **go-playground/validator** | 🆕 Add for input validation |
| Testing | **testify** | 🆕 Add comprehensive tests |

#### Deployment Options
1. **Web App:** Host Go server + React SPA (Nginx/Caddy)
2. **Desktop App:** Electron wrapper (like VS Code)
3. **Hybrid:** Tauri (Rust + WebView) - lighter than Electron

---

### Key Features

#### Phase 1: Foundation (MVP)
- [ ] Asset browser (WAD file tree)
- [ ] 3D viewport with orbit controls
- [ ] Mesh rendering (static + skinned)
- [ ] Material preview
- [ ] Texture viewer
- [ ] Skeleton visualization
- [ ] Basic transform gizmos (translate only)
- [ ] Properties panel (read-only)
- [ ] Export to GLTF/FBX

#### Phase 2: Editing
- [ ] Transform editing (translate, rotate, scale)
- [ ] Instance manipulation (add, delete, duplicate)
- [ ] Material editing (colors, flags)
- [ ] Texture replacement
- [ ] Undo/Redo system
- [ ] Scene hierarchy editing
- [ ] Save changes back to WAD
- [ ] Project system (metadata)

#### Phase 3: Animation
- [ ] Animation playback
- [ ] Timeline/Dope sheet
- [ ] Keyframe editing
- [ ] Animation import/export
- [ ] Bone manipulation
- [ ] IK constraints (optional)

#### Phase 4: Advanced
- [ ] Mesh editing (vertex manipulation)
- [ ] LOD management
- [ ] Collision editing
- [ ] Script editing
- [ ] TWK parameter editor (already exists!)
- [ ] Multi-user collaboration (WebSocket)
- [ ] Auto-save + version history
- [ ] Plugin system

---

### UI Mockup (Conceptual)

```
┌──────────────────────────────────────────────────────────────────────┐
│ File  Edit  View  Tools  Help                        [User] [Settings]│
├─────┬────────────────────────────────────────────────────────────┬────┤
│     │                                                            │    │
│  A  │                                                            │ P  │
│  S  │                                                            │ R  │
│  S  │                    3D VIEWPORT                             │ O  │
│  E  │                   (Three.js)                               │ P  │
│  T  │                                                            │ E  │
│     │              [Translate] [Rotate] [Scale]                  │ R  │
│  B  │                                                            │ T  │
│  R  │                                                            │ I  │
│  O  │                                                            │ E  │
│  W  │                                                            │ S  │
│  S  │                                                            │    │
│  E  │                                                            │ P  │
│  R  │                                                            │ A  │
│     │                                                            │ N  │
│     │                                                            │ E  │
│     │                                                            │ L  │
├─────┴─────────────┬──────────────────────────────────────────────┴────┤
│ SCENE HIERARCHY   │          TIMELINE / ANIMATION                     │
│ □ Scene           │  ├──┼──┼──┼──┼──┼──┼──┼──┼──┼──┼──┤              │
│   ├─□ Instance1   │  0s    1s    2s    3s    4s    5s                │
│   │  └─□ Model    │                                                   │
│   └─□ Instance2   │  ◆─────◆────────────◆──────────◆ (Position.X)    │
└───────────────────┴───────────────────────────────────────────────────┘
```

**Panels:**
- **Asset Browser (Left):** Tree view of WAD files/nodes
- **Viewport (Center):** Three.js canvas with gizmos
- **Properties (Right):** Context-sensitive properties
- **Scene Hierarchy (Bottom-left):** Current scene tree
- **Timeline (Bottom-center):** Animation editing

---

## Implementation Roadmap

### Phase 1: Foundation (4-6 weeks)

#### Week 1-2: Project Setup & Basic UI
- [ ] Initialize React + TypeScript + Vite project
- [ ] Set up Three.js with basic scene
- [ ] Implement layout with resizable panels
- [ ] Create asset browser component
- [ ] Set up routing (React Router)
- [ ] Implement API client layer
- [ ] Add loading states + error handling

**Deliverable:** Empty editor with panels and file browser

#### Week 3-4: 3D Rendering
- [ ] Load MESH data via API and render in Three.js
- [ ] Implement material system (convert MAT → THREE.Material)
- [ ] Texture loading and caching
- [ ] Skeleton visualization (lines + joints)
- [ ] Camera controls (OrbitControls)
- [ ] Selection system (raycaster)
- [ ] Basic lighting (ambient + directional)

**Deliverable:** View GoW scenes in 3D (read-only)

#### Week 5-6: Properties & State
- [ ] Properties panel (display mesh/material/skeleton data)
- [ ] Scene hierarchy tree
- [ ] State management setup (Zustand)
- [ ] Selection state (single + multi-select)
- [ ] View modes (wireframe, shaded, textured)
- [ ] Debug overlays (skeleton, normals, UVs)

**Deliverable:** Full read-only viewer with inspector

---

### Phase 2: Editing (6-8 weeks)

#### Week 7-9: Transform Editing
- [ ] Implement TransformControls (translate gizmo)
- [ ] Update instance position via API
- [ ] Implement rotation gizmo
- [ ] Implement scale gizmo
- [ ] Snap to grid option
- [ ] Transform in local/world space
- [ ] Numeric input for precise values

**Backend:**
- [ ] Implement INST serialization (write binary)
- [ ] POST /api/instance/{file}/{id} endpoint
- [ ] Validate instance data

**Deliverable:** Move objects around and save

#### Week 10-12: Undo/Redo + Project System
- [ ] Command pattern implementation
- [ ] Undo/Redo stack with Zustand
- [ ] Keyboard shortcuts (Ctrl+Z, Ctrl+Y)
- [ ] Project metadata (JSON file)
- [ ] Auto-save functionality
- [ ] Export project state
- [ ] Import project state

**Backend:**
- [ ] Project API endpoints
- [ ] File watcher for external changes

**Deliverable:** Robust editing with undo/redo

#### Week 13-14: Material & Texture Editing
- [ ] Material editor UI (color pickers, flags)
- [ ] Texture replacement (drag-drop)
- [ ] Texture preview (mipmap levels)
- [ ] Material preview sphere
- [ ] Layer management (add/remove layers)

**Backend:**
- [ ] MAT serialization
- [ ] TXR/GFX serialization (already exists!)
- [ ] POST /api/material/{file}/{id}

**Deliverable:** Edit materials and textures

---

### Phase 3: Animation (4-6 weeks)

#### Week 15-17: Animation Playback
- [ ] Load ANM data and convert to THREE.AnimationClip
- [ ] Animation mixer integration
- [ ] Play/Pause/Stop controls
- [ ] Timeline scrubbing
- [ ] Animation list (select clips)
- [ ] Speed control
- [ ] Loop modes

**Backend:**
- [ ] ANM → GLTF animation converter
- [ ] Handle all 12 animation types

**Deliverable:** Play animations in editor

#### Week 18-20: Animation Editing
- [ ] Timeline/Dope sheet UI
- [ ] Keyframe visualization
- [ ] Add/Delete/Move keyframes
- [ ] Interpolation mode (linear, bezier)
- [ ] Bone selection in viewport
- [ ] Record mode (manipulate bones → keyframes)

**Backend:**
- [ ] ANM serialization (complex!)
- [ ] Keyframe optimizer
- [ ] Fixed-point conversion (Q.14)

**Deliverable:** Edit and create animations

---

### Phase 4: Advanced Features (8-12 weeks)

#### Mesh Editing
- [ ] Vertex selection mode
- [ ] Vertex manipulation (translate)
- [ ] Edge/Face selection
- [ ] Extrude, bevel, inset (basic)
- [ ] Weight painting (bone weights)
- [ ] UV editor

**Backend:**
- [ ] MESH serialization (most complex!)
- [ ] DMA packet compiler for PS2
- [ ] Triangle strip optimizer
- [ ] Validation (manifold, weights sum to 1.0)

#### Scene Composition
- [ ] Add new instances (duplicate existing)
- [ ] Delete instances
- [ ] Import GLTF models (convert to GoW format)
- [ ] Batch operations
- [ ] Prefab system

#### Collaboration (Optional)
- [ ] WebSocket real-time updates
- [ ] Multi-user editing (operational transform or CRDTs)
- [ ] Change tracking
- [ ] Conflict resolution

#### Optimization Tools
- [ ] LOD generation (mesh decimation)
- [ ] Texture compression
- [ ] Mesh merging (reduce draw calls)
- [ ] Statistics panel (poly count, texture memory)

---

### Total Timeline: ~6 months for full-featured editor

**MVP (Phase 1+2):** 10-14 weeks
**With Animation (Phase 1+2+3):** 14-20 weeks
**Complete (All Phases):** 22-32 weeks

---

## Technical Specifications

### Data Flow: Edit Operation

```
┌──────────────────────────────────────────────────────────────┐
│ 1. User Action (Frontend)                                    │
│    └─> TransformControls dragged                             │
│        └─> Update local Three.js object                      │
│            └─> Dispatch action to Zustand store              │
└────────────────────────┬─────────────────────────────────────┘
                         │
┌────────────────────────▼─────────────────────────────────────┐
│ 2. Command Pattern (Frontend)                                │
│    └─> Create TransformCommand                               │
│        ├─> Store old position                                │
│        ├─> Store new position                                │
│        └─> Push to undo stack                                │
└────────────────────────┬─────────────────────────────────────┘
                         │
┌────────────────────────▼─────────────────────────────────────┐
│ 3. API Call (Frontend)                                       │
│    └─> POST /api/instance/{file}/{id}                        │
│        Body: { position: [x, y, z], rotation: [...], ... }   │
└────────────────────────┬─────────────────────────────────────┘
                         │ HTTP
┌────────────────────────▼─────────────────────────────────────┐
│ 4. Handler (Backend)                                         │
│    └─> Parse request                                         │
│        └─> Validate data                                     │
│            └─> Call serializer                               │
└────────────────────────┬─────────────────────────────────────┘
                         │
┌────────────────────────▼─────────────────────────────────────┐
│ 5. Serialization (Backend)                                   │
│    └─> Convert JSON → Go struct                              │
│        └─> Convert struct → binary                           │
│            ├─> Write header (magic, size, name)              │
│            ├─> Write data (little-endian)                    │
│            └─> Calculate offsets                             │
└────────────────────────┬─────────────────────────────────────┘
                         │
┌────────────────────────▼─────────────────────────────────────┐
│ 6. File Update (Backend)                                     │
│    └─> Open WAD file                                         │
│        └─> Locate tag by ID                                  │
│            └─> Replace tag data                              │
│                └─> Update WAD header                         │
│                    └─> Flush to disk                         │
└────────────────────────┬─────────────────────────────────────┘
                         │
┌────────────────────────▼─────────────────────────────────────┐
│ 7. Response (Backend)                                        │
│    └─> Return success + updated data                         │
└────────────────────────┬─────────────────────────────────────┘
                         │ HTTP
┌────────────────────────▼─────────────────────────────────────┐
│ 8. Update UI (Frontend)                                      │
│    └─> Update Zustand store                                  │
│        └─> Trigger re-render                                 │
│            └─> Show success notification                     │
└──────────────────────────────────────────────────────────────┘
```

---

### Serialization Examples

#### INST (Instance) Serializer

```go
// pack/wad/inst/marshal.go (NEW FILE)
package inst

import (
    "bytes"
    "encoding/binary"
    "github.com/mogaika/god_of_war_browser/utils"
)

// SerializeGow1 converts Instance struct back to binary
func (inst *Instance) SerializeGow1() ([]byte, error) {
    buf := new(bytes.Buffer)

    // Write object name (0x18 bytes, null-padded)
    nameBytes := make([]byte, 0x18)
    copy(nameBytes, inst.Object)
    buf.Write(nameBytes)

    // Write ID (uint16)
    binary.Write(buf, binary.LittleEndian, inst.Id)

    // Padding
    binary.Write(buf, binary.LittleEndian, uint16(0))

    // Write Position1 (Vec4)
    binary.Write(buf, binary.LittleEndian, inst.Position1)

    // Write Rotation (Vec4 - Euler XYZ + Scale W)
    binary.Write(buf, binary.LittleEndian, inst.Rotation)

    // Write Position2 (Vec4)
    binary.Write(buf, binary.LittleEndian, inst.Position2)

    return buf.Bytes(), nil
}

// Validate checks if instance data is valid
func (inst *Instance) Validate() error {
    if inst.Object == "" {
        return fmt.Errorf("object reference cannot be empty")
    }
    if inst.Position1[3] < 0 || inst.Position1[3] > 1 {
        return fmt.Errorf("position1.w must be 0-1, got %f", inst.Position1[3])
    }
    // More validation...
    return nil
}
```

#### API Handler

```go
// web/handler_instance.go (NEW FILE)
package web

func (ws *Service) HandleUpdateInstance(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    fileName := vars["file"]
    nodeIdStr := vars["id"]

    // Parse node ID
    nodeId, err := strconv.ParseUint(nodeIdStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid node ID", http.StatusBadRequest)
        return
    }

    // Load WAD
    wad, err := ws.wadStore.Load(fileName)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    // Parse request body
    var instData inst.Instance
    if err := json.NewDecoder(r.Body).Decode(&instData); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Validate
    if err := instData.Validate(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Serialize
    binaryData, err := instData.SerializeGow1()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Update WAD tag
    node := wad.GetNode(wad.NodeId(nodeId))
    if node == nil {
        http.Error(w, "Node not found", http.StatusNotFound)
        return
    }

    node.Tag.Data = binaryData
    node.Tag.Size = uint32(len(binaryData))

    // Write back to file
    if err := wad.Save(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Return success
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "nodeId": nodeId,
        "size": len(binaryData),
    })
}
```

---

### Undo/Redo System

```typescript
// frontend/src/store/historyStore.ts
import create from 'zustand';
import { immer } from 'zustand/middleware/immer';

interface Command {
  execute(): void;
  undo(): void;
  redo(): void;
  description: string;
}

class TransformCommand implements Command {
  constructor(
    private objectId: string,
    private oldTransform: Transform,
    private newTransform: Transform,
    private applyFn: (id: string, transform: Transform) => void
  ) {}

  execute() {
    this.applyFn(this.objectId, this.newTransform);
  }

  undo() {
    this.applyFn(this.objectId, this.oldTransform);
  }

  redo() {
    this.execute();
  }

  get description() {
    return `Transform ${this.objectId}`;
  }
}

interface HistoryState {
  undoStack: Command[];
  redoStack: Command[];
  execute: (cmd: Command) => void;
  undo: () => void;
  redo: () => void;
  canUndo: () => boolean;
  canRedo: () => boolean;
}

export const useHistory = create<HistoryState>()(
  immer((set, get) => ({
    undoStack: [],
    redoStack: [],

    execute: (cmd) => {
      cmd.execute();
      set((state) => {
        state.undoStack.push(cmd);
        state.redoStack = []; // Clear redo stack
      });
    },

    undo: () => {
      const { undoStack, redoStack } = get();
      if (undoStack.length === 0) return;

      const cmd = undoStack[undoStack.length - 1];
      cmd.undo();

      set((state) => {
        state.undoStack.pop();
        state.redoStack.push(cmd);
      });
    },

    redo: () => {
      const { redoStack } = get();
      if (redoStack.length === 0) return;

      const cmd = redoStack[redoStack.length - 1];
      cmd.redo();

      set((state) => {
        state.redoStack.pop();
        state.undoStack.push(cmd);
      });
    },

    canUndo: () => get().undoStack.length > 0,
    canRedo: () => get().redoStack.length > 0,
  }))
);
```

---

## Risk Analysis

### Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| **PS2 DMA packet compilation complexity** | High | High | Start with GoW2018 (simpler format), defer PS2 mesh writing to Phase 4 |
| **Animation serialization bugs** | Medium | High | Extensive testing with original files, diff tool to compare binary output |
| **Browser performance (large scenes)** | Medium | Medium | LOD system, frustum culling, instancing, Web Workers for parsing |
| **Fixed-point arithmetic precision** | Low | Medium | Use existing parsing code as reference, unit tests with known values |
| **File corruption during editing** | Low | Critical | Always backup original, implement validation, atomic writes |
| **WebGL compatibility** | Low | Low | Use Three.js (handles fallbacks), target WebGL 2.0 with 1.0 fallback |

### Project Risks

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| **Scope creep** | High | Medium | Follow phased approach, MVP first, resist feature requests early |
| **Underestimated complexity** | Medium | High | Build PS2 mesh serialization POC early (week 3-4) to validate feasibility |
| **Team size (solo dev?)** | N/A | High | Focus on MVP, open-source for community contributions |
| **User adoption** | Medium | Low | Engage GoW modding community early, gather feedback |
| **Maintenance burden** | Medium | Medium | Good documentation, clean architecture, automated tests |

---

## Alternatives Considered (Summary)

### Why Not Godot/Unity/Blender?

**GLTF Conversion Loss:**
The fundamental problem with using external editors is **format mismatch**. God of War formats contain:
- PS2-specific data (DMA packets, swizzled textures)
- Multiple blend color layers
- 12 animation types (not just skeletal)
- Game-specific flags and metadata

Converting to GLTF loses this data. Round-tripping (GoW → GLTF → Edit → GLTF → GoW) would require:
1. Custom GLTF extensions (complex, fragile)
2. Sidecar files for extra data (messy)
3. Validation to prevent data loss (difficult)

**Better Approach:**
- Keep Go backend with direct format access
- Build custom editor that understands GoW formats natively
- Still support GLTF export for users who want external tools

---

## Conclusion

### Recommended Approach: Custom Web-Based Editor

**Why:**
1. ✅ **Leverage existing work:** 90% of parsing done
2. ✅ **Direct format control:** No lossy conversions
3. ✅ **Modern tech stack:** React + Three.js are mature and powerful
4. ✅ **Cross-platform:** Runs anywhere
5. ✅ **Incremental development:** MVP in 3 months, full editor in 6 months
6. ✅ **Community friendly:** Easy to contribute (TypeScript/Go)

**What to Build:**
- **Phase 1 (MVP):** Read-only viewer with modern UI (3 months)
- **Phase 2:** Transform editing + materials (2 months)
- **Phase 3:** Animation editing (1.5 months)
- **Phase 4:** Advanced features (as needed)

**Key Success Factors:**
1. Start with simplest formats (INST, MAT, TXR)
2. Build comprehensive tests (compare original vs. re-serialized binaries)
3. Engage modding community early
4. Maintain backward compatibility (read old saves)
5. Document binary formats (for future maintainers)

---

## Next Steps

1. **Validate feasibility:** Build POC for INST serialization (1-2 days)
2. **Set up dev environment:** React + Vite + Three.js boilerplate (1 day)
3. **Implement basic viewer:** Load MESH and render (1 week)
4. **Add transform editing:** Gizmos + INST write endpoint (1 week)
5. **Build undo/redo:** Command pattern + Zustand (3 days)
6. **Iterate:** Add features based on user feedback

---

## Appendix: Useful Resources

### Three.js
- [Three.js Documentation](https://threejs.org/docs/)
- [Three.js Examples](https://threejs.org/examples/)
- [TransformControls](https://threejs.org/docs/#examples/en/controls/TransformControls)
- [GLTFLoader](https://threejs.org/docs/#examples/en/loaders/GLTFLoader)

### React Ecosystem
- [React Three Fiber](https://docs.pmnd.rs/react-three-fiber) (React renderer for Three.js)
- [Drei](https://github.com/pmndrs/drei) (Useful helpers for R3F)
- [Zustand](https://github.com/pmndrs/zustand) (State management)
- [Leva](https://github.com/pmndrs/leva) (GUI controls)

### Game Editor References
- [Unity Editor Architecture](https://docs.unity3d.com/Manual/editor-EditorWindows.html)
- [Godot Engine Docs](https://docs.godotengine.org/)
- [Blender Python API](https://docs.blender.org/api/current/)

### GLTF Specification
- [GLTF 2.0 Spec](https://registry.khronos.org/glTF/specs/2.0/glTF-2.0.html)
- [GLTF Extensions](https://github.com/KhronosGroup/glTF/tree/main/extensions)

---

**Document Version:** 1.0
**Last Updated:** 2025-11-11
**Author:** Claude (AI Assistant)
**Status:** Ready for Review
