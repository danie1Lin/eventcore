import _ from "lodash";

const VueLodash = {
  install(Vue) {
    Vue.prototype.$_ = _;
    if (typeof window !== "undefined" && window.Vue) {
      window.Vue.use(VueLodash);
    }
  }
};

export default VueLodash;
