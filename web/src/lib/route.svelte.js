let path = $state(location.pathname);

window.addEventListener('popstate', () => {
  path = location.pathname;
});

export const route = {
  get path() {
    return path;
  },
  go(next) {
    history.pushState({}, '', next);
    path = next;
  },
};
