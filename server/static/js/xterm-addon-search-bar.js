!(function (e, t) {
  "object" == typeof exports && "undefined" != typeof module
    ? t(exports)
    : "function" == typeof define && define.amd
    ? define(["exports"], t)
    : t((e.SearchBarAddon = {}));
})(this, function (e) {
  "use strict";
  !(function (e, t) {
    void 0 === t && (t = {});
    var a = t.insertAt;
    if (e && "undefined" != typeof document) {
      var i = document.head || document.getElementsByTagName("head")[0],
        n = document.createElement("style");
      (n.type = "text/css"),
        "top" === a && i.firstChild
          ? i.insertBefore(n, i.firstChild)
          : i.appendChild(n),
        n.styleSheet
          ? (n.styleSheet.cssText = e)
          : n.appendChild(document.createTextNode(e));
    }
  })(
    `.xterm-search-bar__addon{position:absolute;max-width:1467px;top:0;right:28px;color:#000;background:#fff;
        padding:5px 10px;box-shadow:0 2px 8px #000;background-color:#252526;z-index:999;display:flex}.xterm-search-bar__addon 
      .search-bar__input{background-color:#3c3c3c;color:#ccc;border:0;margin-bottom:0px;padding:2px;height:20px;width:227px}.xterm-search-bar__addon
      .search-bar__btn{min-width:20px;width:20px;height:20px;display:flex;display:-webkit-flex;flex:initial;background-position:50%;
        margin-left:3px;margin-bottom:0px;background-repeat:no-repeat;background-color:#252526;border:0;cursor:pointer;padding: 0}.xterm-search-bar__addon
      .search-bar__btn:hover{background-color:#666}.xterm-search-bar__addon .search-bar__btn.prev{background-image:url("data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIxNiIgaGVpZ2h0PSIxNiIgZmlsbD0iI2ZmZiI+PHBhdGggZD0iTTUuNCA4YS42LjYgMCAwMS4xNzYtLjQyNGw0LTRhLjU5OC41OTggMCAwMS44NDggMCAuNTk4LjU5OCAwIDAxMCAuODQ4TDYuODQ5IDhsMy41NzUgMy41NzZhLjU5OC41OTggMCAwMTAgLjg0OC41OTguNTk4IDAgMDEtLjg0OCAwbC00LTRBLjYuNiAwIDAxNS40IDgiLz48L3N2Zz4=")}.xterm-search-bar__addon
      .search-bar__btn.next{background-image:url("data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIxNiIgaGVpZ2h0PSIxNiIgZmlsbD0iI2ZmZiI+PHBhdGggZD0iTTEwLjYgOGEuNi42IDAgMDEtLjE3Ni40MjRsLTQgNGEuNTk4LjU5OCAwIDAxLS44NDggMCAuNTk4LjU5OCAwIDAxMC0uODQ4TDkuMTUxIDggNS41NzYgNC40MjRhLjU5OC41OTggMCAwMTAtLjg0OC41OTguNTk4IDAgMDEuODQ4IDBsNCA0QS42LjYgMCAwMTEwLjYgOCIvPjwvc3ZnPg==")}.xterm-search-bar__addon
      .search-bar__btn.close{background-image:url("data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIxMiIgaGVpZ2h0PSIxMiIgZmlsbD0iI2ZmZiI+PHBhdGggZD0iTTcgNmwyLTJhLjcxMS43MTEgMCAwMDAtMSAuNzExLjcxMSAwIDAwLTEgMEw2IDUgNCAzYS43MTEuNzExIDAgMDAtMSAwIC43MTEuNzExIDAgMDAwIDFsMiAyLTIgMmEuNzExLjcxMSAwIDAwMCAxIC43MTEuNzExIDAgMDAxIDBsMi0yIDIgMmEuNzExLjcxMSAwIDAwMSAwIC43MTEuNzExIDAgMDAwLTFMNyA2eiIvPjwvc3ZnPg==")}`
  );
  const t = "xterm-search-bar__addon";
  (e.SearchBarAddon = class {
    constructor(e) {
      (this.options = e || {}),
        this.options &&
          this.options.searchAddon &&
          (this.searchAddon = this.options.searchAddon);
    }
    activate(e) {
      (this.terminal = e), this.searchAddon;
    }
    dispose() {
      this.hidden();
    }
    show() {
      if (!this.terminal || !this.terminal.element) return;
      if (this.searchBarElement)
        return (
          (this.searchBarElement.style.visibility = "visible"),
          void this.searchBarElement.querySelector("input").select()
        );
      this.terminal.element.style.position = "relative";
      const e = document.createElement("div");
      (e.innerHTML =
        `<input type="text" class="search-bar__input" name="search-bar__input"/>
         <button class="search-bar__btn prev"></button>
         <button class="search-bar__btn next"></button>
         <button class="search-bar__btn close"></button>`),
        (e.className = t);
      const a = this.terminal.element.parentElement;
      (this.searchBarElement = e),
/*        ["relative", "absolute", "fixed"].includes(a.style.position) ||
          (a.style.position = "relative"),*/
        a.appendChild(this.searchBarElement),
        this.on(".search-bar__btn.close", "click", () => {
          this.hidden();
        }),
        this.on(".search-bar__btn.next", "click", () => {
          this.searchAddon.findNext(this.searchKey, { incremental: !1 });
        }),
        this.on(".search-bar__btn.prev", "click", () => {
          this.searchAddon.findPrevious(this.searchKey, { incremental: !1 });
        }),
        this.on(".search-bar__input", "keyup", (e) => {
          (this.searchKey = e.target.value),
            this.searchAddon.findNext(this.searchKey, {
              incremental: "Enter" !== e.key,
            });
        }),
        this.searchBarElement.querySelector("input").select();
    }
    hidden() {
      this.searchBarElement &&
        this.terminal.element.parentElement &&
        (this.searchBarElement.style.visibility = "hidden");
    }
    on(e, t, a) {
      const i = this.terminal.element.parentElement;
      i.addEventListener(t, (t) => {
        let n = t.target;
        for (; n !== document.querySelector(e); ) {
          if (n === i) {
            n = null;
            break;
          }
          n = n.parentElement;
        }
        n === document.querySelector(e) &&
          (a.call(this, t), t.stopPropagation());
      });
    }
    addNewStyle(e) {
      let a = document.getElementById(t);
      a ||
        (((a = document.createElement("style")).type = "text/css"),
        (a.id = t),
        document.getElementsByTagName("head")[0].appendChild(a)),
        a.appendChild(document.createTextNode(e));
    }
  }),
    Object.defineProperty(e, "__esModule", { value: !0 });
});
