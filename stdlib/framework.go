package stdlib

// FrameworkRuntime contains the JavaScript runtime for the Quill reactive UI framework.
// It provides reactive state management, virtual DOM diffing, event handling,
// two-way binding, conditional rendering, list rendering, component composition,
// CSS scoping, and lifecycle hooks.
const FrameworkRuntime = `
// Quill Reactive UI Framework Runtime

(function(global) {
  "use strict";

  // --- Virtual DOM ---

  function h(tag, props) {
    var children = [];
    for (var i = 2; i < arguments.length; i++) {
      var child = arguments[i];
      if (Array.isArray(child)) {
        for (var j = 0; j < child.length; j++) {
          children.push(child[j]);
        }
      } else if (child != null && child !== false && child !== undefined) {
        children.push(child);
      }
    }
    return { tag: tag, props: props || {}, children: children };
  }

  // --- Diffing & Patching ---

  function createDOMNode(vnode) {
    if (typeof vnode === "string" || typeof vnode === "number") {
      return document.createTextNode(String(vnode));
    }
    if (!vnode || !vnode.tag) {
      return document.createTextNode("");
    }
    var el = document.createElement(vnode.tag);
    applyProps(el, {}, vnode.props);
    for (var i = 0; i < vnode.children.length; i++) {
      el.appendChild(createDOMNode(vnode.children[i]));
    }
    return el;
  }

  function applyProps(el, oldProps, newProps) {
    // Remove old props
    for (var key in oldProps) {
      if (!(key in newProps)) {
        if (key.slice(0, 2) === "on") {
          el.removeEventListener(key.slice(2).toLowerCase(), oldProps[key]);
        } else if (key === "className") {
          el.removeAttribute("class");
        } else if (key === "style" && typeof oldProps[key] === "object") {
          el.removeAttribute("style");
        } else {
          el.removeAttribute(key);
        }
      }
    }
    // Set new props
    for (var key in newProps) {
      var val = newProps[key];
      if (key.slice(0, 2) === "on") {
        var evtName = key.slice(2).toLowerCase();
        if (oldProps[key]) {
          el.removeEventListener(evtName, oldProps[key]);
        }
        el.addEventListener(evtName, val);
      } else if (key === "className") {
        el.setAttribute("class", val);
      } else if (key === "style" && typeof val === "object") {
        for (var sk in val) {
          el.style[sk] = val[sk];
        }
      } else if (key === "value") {
        el.value = val;
      } else if (key === "checked") {
        el.checked = val;
      } else {
        el.setAttribute(key, val);
      }
    }
  }

  function diff(parent, oldVNode, newVNode, index) {
    index = index || 0;
    var existing = parent.childNodes[index];

    // No old node - append new
    if (oldVNode == null) {
      parent.appendChild(createDOMNode(newVNode));
      return;
    }

    // No new node - remove old
    if (newVNode == null) {
      if (existing) parent.removeChild(existing);
      return;
    }

    // Text nodes
    if ((typeof oldVNode === "string" || typeof oldVNode === "number") &&
        (typeof newVNode === "string" || typeof newVNode === "number")) {
      if (String(oldVNode) !== String(newVNode) && existing) {
        existing.textContent = String(newVNode);
      }
      return;
    }

    // Type changed or tag changed - replace entirely
    if (typeof oldVNode !== typeof newVNode ||
        (oldVNode.tag && newVNode.tag && oldVNode.tag !== newVNode.tag)) {
      if (existing) {
        parent.replaceChild(createDOMNode(newVNode), existing);
      }
      return;
    }

    // Same tag - update props and diff children
    if (newVNode.tag && existing) {
      applyProps(existing, oldVNode.props || {}, newVNode.props || {});

      var oldChildren = oldVNode.children || [];
      var newChildren = newVNode.children || [];
      var maxLen = Math.max(oldChildren.length, newChildren.length);

      // Diff children in reverse so removals don't shift indices
      for (var i = maxLen - 1; i >= 0; i--) {
        diff(existing, oldChildren[i] || null, newChildren[i] || null, i);
      }

      // Remove extra old children
      while (existing.childNodes.length > newChildren.length) {
        existing.removeChild(existing.lastChild);
      }
    }
  }

  // --- Component Class ---

  var componentId = 0;

  function QuillComponent(definition) {
    var self = this;
    self.__id = "qc-" + (componentId++);
    self.__definition = definition;
    self.__mounted = false;
    self.__rootEl = null;
    self.__oldVdom = null;
    self.__rendering = false;
    self.__pendingRender = false;
    self.__onMountCallbacks = [];
    self.__onDestroyCallbacks = [];
    self.__childComponents = [];

    // Initialize state with reactive proxy
    var stateData = {};
    if (definition.initialState) {
      var init = definition.initialState();
      for (var k in init) {
        stateData[k] = init[k];
      }
    }

    self.state = new Proxy(stateData, {
      set: function(target, prop, value) {
        target[prop] = value;
        if (self.__mounted && !self.__rendering) {
          self.__scheduleRender();
        }
        return true;
      }
    });

    // Bind methods
    if (definition.methods) {
      var methods = definition.methods(self);
      for (var mname in methods) {
        self[mname] = methods[mname].bind(self);
      }
    }

    // Lifecycle: onMount
    if (definition.onMount) {
      self.__onMountCallbacks.push(definition.onMount.bind(self));
    }

    // Lifecycle: onDestroy
    if (definition.onDestroy) {
      self.__onDestroyCallbacks.push(definition.onDestroy.bind(self));
    }
  }

  QuillComponent.prototype.__scheduleRender = function() {
    if (this.__pendingRender) return;
    this.__pendingRender = true;
    var self = this;
    requestAnimationFrame(function() {
      self.__pendingRender = false;
      self.__update();
    });
  };

  QuillComponent.prototype.__update = function() {
    if (!this.__mounted || !this.__rootEl) return;
    this.__rendering = true;
    var newVdom = this.__definition.render(this);
    if (this.__oldVdom) {
      diff(this.__rootEl, this.__oldVdom, newVdom, 0);
    } else {
      this.__rootEl.innerHTML = "";
      this.__rootEl.appendChild(createDOMNode(newVdom));
    }
    this.__oldVdom = newVdom;
    this.__rendering = false;
  };

  QuillComponent.prototype.mount = function(selector) {
    var el = typeof selector === "string" ? document.querySelector(selector) : selector;
    if (!el) {
      throw new Error("Quill: Could not find mount target: " + selector);
    }
    this.__rootEl = el;

    // Add scoped attribute
    el.setAttribute("data-quill-" + this.__id, "");

    // Initial render
    this.__rendering = true;
    var vdom = this.__definition.render(this);
    el.innerHTML = "";
    el.appendChild(createDOMNode(vdom));
    this.__oldVdom = vdom;
    this.__rendering = false;
    this.__mounted = true;

    // Fire onMount callbacks
    for (var i = 0; i < this.__onMountCallbacks.length; i++) {
      this.__onMountCallbacks[i]();
    }

    return this;
  };

  QuillComponent.prototype.destroy = function() {
    // Fire onDestroy callbacks
    for (var i = 0; i < this.__onDestroyCallbacks.length; i++) {
      this.__onDestroyCallbacks[i]();
    }
    // Destroy child components
    for (var i = 0; i < this.__childComponents.length; i++) {
      this.__childComponents[i].destroy();
    }
    this.__mounted = false;
    if (this.__rootEl) {
      this.__rootEl.innerHTML = "";
    }
  };

  // --- Mount helper ---

  function __quill_mount(ComponentDef, selector) {
    var comp = new QuillComponent(ComponentDef);
    comp.mount(selector);
    return comp;
  }

  // --- Expose to global ---
  global.h = h;
  global.QuillComponent = QuillComponent;
  global.__quill_mount = __quill_mount;

  // Standard library helpers for components
  global.push = function(arr, item) { arr.push(item); return arr; };
  global.pop = function(arr) { return arr.pop(); };
  global.len = function(arr) { return arr.length; };
  global.removeAt = function(arr, i) { arr.splice(i, 1); return arr; };
  global.insertAt = function(arr, i, item) { arr.splice(i, 0, item); return arr; };

})(typeof window !== "undefined" ? window : global);
`
