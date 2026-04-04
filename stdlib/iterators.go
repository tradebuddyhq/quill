package stdlib

// GetIteratorRuntime returns the JavaScript runtime for lazy iterators and generators.
func GetIteratorRuntime() string {
	return iteratorRuntime
}

const iteratorRuntime = `
class __QuillLazy {
  constructor(iterable) { this.iterable = iterable; }

  filter(fn) {
    const self = this;
    return new __QuillLazy(function*() {
      for (const item of self.iterable) {
        if (fn(item)) yield item;
      }
    }());
  }

  map(fn) {
    const self = this;
    return new __QuillLazy(function*() {
      for (const item of self.iterable) {
        yield fn(item);
      }
    }());
  }

  take(n) {
    const self = this;
    return new __QuillLazy(function*() {
      let count = 0;
      for (const item of self.iterable) {
        if (count >= n) break;
        yield item;
        count++;
      }
    }());
  }

  skip(n) {
    const self = this;
    return new __QuillLazy(function*() {
      let count = 0;
      for (const item of self.iterable) {
        if (count >= n) yield item;
        count++;
      }
    }());
  }

  takeWhile(fn) {
    const self = this;
    return new __QuillLazy(function*() {
      for (const item of self.iterable) {
        if (!fn(item)) break;
        yield item;
      }
    }());
  }

  skipWhile(fn) {
    const self = this;
    return new __QuillLazy(function*() {
      let skipping = true;
      for (const item of self.iterable) {
        if (skipping && fn(item)) continue;
        skipping = false;
        yield item;
      }
    }());
  }

  zip(other) {
    const self = this;
    return new __QuillLazy(function*() {
      const iter1 = self.iterable[Symbol.iterator]();
      const iter2 = other[Symbol.iterator] ? other[Symbol.iterator]() : other.iterable[Symbol.iterator]();
      while (true) {
        const a = iter1.next(), b = iter2.next();
        if (a.done || b.done) break;
        yield [a.value, b.value];
      }
    }());
  }

  enumerate() {
    const self = this;
    return new __QuillLazy(function*() {
      let i = 0;
      for (const item of self.iterable) {
        yield [i, item];
        i++;
      }
    }());
  }

  flatten() {
    const self = this;
    return new __QuillLazy(function*() {
      for (const item of self.iterable) {
        if (item[Symbol.iterator]) yield* item;
        else yield item;
      }
    }());
  }

  collect() {
    return [...this.iterable];
  }

  reduce(fn, init) {
    let acc = init;
    for (const item of this.iterable) acc = fn(acc, item);
    return acc;
  }

  forEach(fn) {
    for (const item of this.iterable) fn(item);
  }

  count() {
    let n = 0;
    for (const _ of this.iterable) n++;
    return n;
  }

  first() {
    for (const item of this.iterable) return item;
    return null;
  }

  last() {
    let last = null;
    for (const item of this.iterable) last = item;
    return last;
  }

  any(fn) {
    for (const item of this.iterable) if (fn(item)) return true;
    return false;
  }

  every(fn) {
    for (const item of this.iterable) if (!fn(item)) return false;
    return true;
  }

  [Symbol.iterator]() {
    return this.iterable[Symbol.iterator]();
  }
}

function __quill_lazy(iterable) {
  return new __QuillLazy(iterable);
}

function* __quill_range(start, end, step = 1) {
  for (let i = start; i < end; i += step) yield i;
}
`
