import $ from 'jquery';
import {hideElem, showElem} from '../utils/dom.ts';

export function initRepoCloneLink() {
  const repoCloneSSH = document.getElementById('repo-clone-ssh');
  const repoCloneHTTPS = document.getElementById('repo-clone-https');
  const inputLink = document.getElementById('repo-clone-url');

  if ((!repoCloneSSH && !repoCloneHTTPS) || !inputLink) {
    return;
  }

  repoCloneSSH?.addEventListener('click', () => {
    localStorage.setItem('repo-clone-protocol', 'ssh');
    window.updateCloneStates();
  });
  repoCloneHTTPS?.addEventListener('click', () => {
    localStorage.setItem('repo-clone-protocol', 'https');
    window.updateCloneStates();
  });

  inputLink.addEventListener('focus', () => {
    inputLink.select();
  });
}

export function initRepoCommonBranchOrTagDropdown(selector) {
  $(selector).each(function () {
    const $dropdown = $(this);
    $dropdown.find('.branch-tag-item').on('click', function () {
      hideElem($dropdown.find('.scrolling.reference-list-menu'));
      showElem($($(this).data('target')));
      return false;
    });
  });
}

export function initRepoCommonFilterSearchDropdown(selector) {
  const $dropdown = $(selector);
  if (!$dropdown.length) return;

  $dropdown.dropdown({
    fullTextSearch: 'exact',
    selectOnKeydown: false,
    onChange(_text, _value, $choice) {
      if ($choice[0].getAttribute('data-url')) {
        window.location.href = $choice[0].getAttribute('data-url');
      }
    },
    message: {noResults: $dropdown[0].getAttribute('data-no-results')},
  });
}
