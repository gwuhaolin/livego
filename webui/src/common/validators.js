const checked = (value, options) => {
  if (value !== true) {
    return options.message || 'must be checked';
  }
};

export default {
  checked
};
