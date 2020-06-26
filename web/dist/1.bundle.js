(window["webpackJsonp"] = window["webpackJsonp"] || []).push([[1],{

/***/ "./src/event_center_client.js":
/*!************************************!*\
  !*** ./src/event_center_client.js ***!
  \************************************/
/*! exports provided: default */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
eval("__webpack_require__.r(__webpack_exports__);\n/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, \"default\", function() { return _default; });\nfunction asyncGeneratorStep(gen, resolve, reject, _next, _throw, key, arg) { try { var info = gen[key](arg); var value = info.value; } catch (error) { reject(error); return; } if (info.done) { resolve(value); } else { Promise.resolve(value).then(_next, _throw); } }\n\nfunction _asyncToGenerator(fn) { return function () { var self = this, args = arguments; return new Promise(function (resolve, reject) { var gen = fn.apply(self, args); function _next(value) { asyncGeneratorStep(gen, resolve, reject, _next, _throw, \"next\", value); } function _throw(err) { asyncGeneratorStep(gen, resolve, reject, _next, _throw, \"throw\", err); } _next(undefined); }); }; }\n\nfunction _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError(\"Cannot call a class as a function\"); } }\n\nfunction _defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if (\"value\" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } }\n\nfunction _createClass(Constructor, protoProps, staticProps) { if (protoProps) _defineProperties(Constructor.prototype, protoProps); if (staticProps) _defineProperties(Constructor, staticProps); return Constructor; }\n\nvar axios = __webpack_require__(/*! axios */ \"../../node_modules/axios/index.js\")[\"default\"];\n\nvar _ = __webpack_require__(/*! lodash */ \"./node_modules/lodash/lodash.js\");\n\nvar _default = /*#__PURE__*/function () {\n  function _default(host, port) {\n    _classCallCheck(this, _default);\n\n    this.clientID = null;\n    this.ws = null;\n    this.host = host;\n    this.port = port;\n    this.eventFuncs = new Map();\n  }\n\n  _createClass(_default, [{\n    key: \"subscript\",\n    value: function () {\n      var _subscript = _asyncToGenerator( /*#__PURE__*/regeneratorRuntime.mark(function _callee(event) {\n        var _len,\n            funcs,\n            _key,\n            _fs,\n            data,\n            _args = arguments;\n\n        return regeneratorRuntime.wrap(function _callee$(_context) {\n          while (1) {\n            switch (_context.prev = _context.next) {\n              case 0:\n                fs = this.eventFuncs.get(event.type);\n\n                for (_len = _args.length, funcs = new Array(_len > 1 ? _len - 1 : 0), _key = 1; _key < _len; _key++) {\n                  funcs[_key - 1] = _args[_key];\n                }\n\n                if (fs && _.isArray(fs)) {\n                  (_fs = fs).append.apply(_fs, funcs);\n                } else {\n                  this.eventFuncs.set(event.type, funcs);\n                }\n\n                data = {\n                  eventType: event.type\n                };\n\n                _.merge(data, {\n                  clientID: this.clientID\n                });\n\n                _context.next = 7;\n                return axios({\n                  method: 'POST',\n                  url: \"http://\".concat(this.host, \":\").concat(this.port, \"/subscript\"),\n                  contentType: 'application/json',\n                  data: JSON.stringify(data)\n                });\n\n              case 7:\n                result = _context.sent;\n\n                if (result.data.success) {\n                  this.clientID = result.data.clientID;\n                }\n\n              case 9:\n              case \"end\":\n                return _context.stop();\n            }\n          }\n        }, _callee, this);\n      }));\n\n      function subscript(_x) {\n        return _subscript.apply(this, arguments);\n      }\n\n      return subscript;\n    }()\n  }, {\n    key: \"receive\",\n    value: function receive(event) {\n      fs = this.eventFuncs.get(event.type);\n\n      if (_.isArray(fs)) {\n        fs.forEach(function (f) {\n          f(event);\n        });\n      }\n    }\n  }, {\n    key: \"eventTunnel\",\n    value: function eventTunnel() {\n      var _this = this;\n\n      if (_.isEmpty(this.clientID)) {\n        throw 'client Id empty';\n      }\n\n      if (!_.isEmpty(this.ws)) {\n        throw 'ws is already connected';\n      }\n\n      this.ws = new WebSocket(\"ws://\".concat(this.host, \":\").concat(this.port, \"/event_tunnel\"), this.port);\n\n      this.ws.onopen = function () {\n        _this.ws.send(JSON.stringify({\n          client_id: _this.clientID\n        }));\n\n        cnt = 1;\n        setInterval(function () {\n          cnt++;\n\n          _this.ws.send(JSON.stringify({\n            type: 'main.EventTest',\n            message: cnt.toString()\n          }));\n        }, 1000);\n      };\n\n      this.ws.onmessage = function incoming(data) {\n        console.log('receive:', data);\n        this.receive(event);\n      };\n    }\n  }]);\n\n  return _default;\n}();\n\n\n\n//# sourceURL=webpack:///./src/event_center_client.js?");

/***/ })

}]);