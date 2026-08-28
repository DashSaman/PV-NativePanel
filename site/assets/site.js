(() => {
  'use strict';

  const dataPath = '/data/articles.json';
  const menuButton = document.querySelector('.menu-button');
  const nav = document.querySelector('#primary-nav');
  const status = document.querySelector('#cache-status');
  const updated = document.querySelector('#last-updated');
  const today = document.querySelector('#today-label');

  const setToday = () => {
    if (!today) return;
    try {
      const label = new Intl.DateTimeFormat('fa-IR', {
        weekday: 'long', year: 'numeric', month: 'long', day: 'numeric'
      }).format(new Date());
      today.textContent = `${label} · مرور تازه‌ترین خبرها`;
    } catch (_) {
      today.textContent = 'مرور تازه‌ترین خبرها از منابع اصلی';
    }
  };

  const safeUrl = (value) => {
    try {
      const url = new URL(value);
      return url.protocol === 'https:' ? url.href : '#';
    } catch (_) {
      return '#';
    }
  };

  const makeLink = (text, href, className) => {
    const link = document.createElement('a');
    link.textContent = text;
    link.href = safeUrl(href);
    link.target = '_blank';
    link.rel = 'noopener noreferrer';
    if (className) link.className = className;
    return link;
  };

  const meta = (article, source) => {
    const row = document.createElement('div');
    row.className = 'story-meta';
    const category = document.createElement('span');
    category.className = 'category';
    category.textContent = article.category;
    const date = document.createElement('span');
    date.textContent = article.published_label;
    row.append(category, date);
    if (source && source.name) row.title = source.name;
    return row;
  };

  const makeCard = (article, source) => {
    const card = document.createElement('article');
    card.className = 'news-card';

    const image = document.createElement('img');
    image.src = article.image || '/assets/news-fallback.svg';
    image.alt = article.image_alt || 'تصویر جایگزین خبر';
    image.loading = 'lazy';
    image.width = 640;
    image.height = 360;
    image.addEventListener('error', () => { image.src = '/assets/news-fallback.svg'; }, { once: true });

    const body = document.createElement('div');
    body.className = 'news-card__body';
    const title = document.createElement('h3');
    title.append(makeLink(article.title, article.source_url));
    const summary = document.createElement('p');
    summary.textContent = article.summary;
    const footer = document.createElement('div');
    footer.className = 'card-footer';
    const sourceName = document.createElement('span');
    sourceName.textContent = source ? source.domain : 'منبع اصلی';
    footer.append(sourceName, makeLink('اصل خبر', article.source_url));
    body.append(meta(article, source), title, summary, footer);
    card.append(image, body);
    return card;
  };

  const makeStackItem = (article, source) => {
    const item = document.createElement('article');
    const image = document.createElement('img');
    image.src = article.image || '/assets/news-fallback.svg';
    image.alt = article.image_alt || 'تصویر جایگزین خبر';
    image.loading = 'lazy';
    image.width = 240;
    image.height = 150;
    image.addEventListener('error', () => { image.src = '/assets/news-fallback.svg'; }, { once: true });
    const body = document.createElement('div');
    const title = document.createElement('h3');
    title.append(makeLink(article.title, article.source_url));
    const summary = document.createElement('p');
    summary.textContent = article.summary;
    body.append(meta(article, source), title, summary);
    item.append(image, body);
    return item;
  };

  const sourceMap = (payload) => new Map(payload.sources.map((source) => [source.id, source]));

  const replaceList = (id, articles, sources, factory) => {
    const target = document.getElementById(id);
    if (!target || !articles.length) return;
    const fragment = document.createDocumentFragment();
    articles.forEach((article) => fragment.append(factory(article, sources.get(article.source_id))));
    target.replaceChildren(fragment);
  };

  const renderHero = (article, source) => {
    const target = document.getElementById('hero-story');
    if (!target || !article) return;
    const visual = document.createElement('div');
    visual.className = 'hero-story__visual';
    visual.setAttribute('aria-hidden', 'true');
    const image = document.createElement('img');
    image.src = article.image || '/assets/news-fallback.svg';
    image.alt = '';
    image.width = 960;
    image.height = 560;
    visual.append(image);

    const body = document.createElement('div');
    body.className = 'hero-story__body';
    const title = document.createElement('h1');
    title.append(makeLink(article.title, article.source_url));
    const summary = document.createElement('p');
    summary.textContent = article.summary;
    const sourceLine = document.createElement('div');
    sourceLine.className = 'source-line';
    const sourceName = document.createElement('span');
    sourceName.textContent = source ? source.name : 'منبع اصلی';
    sourceLine.append(sourceName, makeLink('مشاهده در منبع اصلی ←', article.source_url));
    body.append(meta(article, source), title, summary, sourceLine);
    target.replaceChildren(visual, body);
  };

  const renderLeader = (articles, sources) => {
    const target = document.getElementById('leader-grid');
    if (!target || !articles.length) return;
    const fragment = document.createDocumentFragment();
    articles.forEach((article) => {
      const item = document.createElement('article');
      const label = document.createElement('span');
      label.textContent = article.category;
      const title = document.createElement('h3');
      title.append(makeLink(article.title, article.source_url));
      const summary = document.createElement('p');
      summary.textContent = article.summary;
      const source = sources.get(article.source_id);
      if (source) item.title = source.name;
      item.append(label, title, summary);
      fragment.append(item);
    });
    target.replaceChildren(fragment);
  };

  const renderBreaking = (article) => {
    if (!article) return;
    const link = document.getElementById('breaking-link');
    if (!link) return;
    link.textContent = article.title;
    link.href = safeUrl(article.source_url);
  };

  const renderPayload = (payload) => {
    if (!payload || !Array.isArray(payload.sources) || !Array.isArray(payload.articles)) throw new Error('invalid local article cache');
    const sources = sourceMap(payload);
    const bySection = (section) => payload.articles.filter((article) => article.section === section);
    renderBreaking(bySection('breaking')[0]);
    renderHero(bySection('hero')[0], sources.get(bySection('hero')[0]?.source_id));
    replaceList('latest-grid', bySection('latest'), sources, makeCard);
    renderLeader(bySection('leader'), sources);
    replaceList('politics-grid', bySection('politics'), sources, makeStackItem);
    replaceList('economy-grid', bySection('economy'), sources, makeStackItem);
    replaceList('international-grid', bySection('international'), sources, makeStackItem);
    if (updated && payload.updated_label) updated.textContent = payload.updated_label;
    if (status) status.textContent = 'نسخه محلی به‌روز';
    document.documentElement.dataset.newsState = 'ready';
  };

  const renderFallback = () => {
    document.documentElement.dataset.newsState = 'fallback';
    if (status) status.textContent = 'نمایش نسخه پایدار ذخیره‌شده';
    if (updated) updated.textContent = 'نسخه پایدار';
  };

  const hydrate = async () => {
    try {
      const response = await fetch(dataPath, { cache: 'no-store', credentials: 'same-origin' });
      if (!response.ok) throw new Error('local cache unavailable');
      renderPayload(await response.json());
    } catch (_) {
      renderFallback();
    }
  };

  if (menuButton && nav) {
    menuButton.addEventListener('click', () => {
      const open = nav.classList.toggle('is-open');
      menuButton.setAttribute('aria-expanded', String(open));
    });
    nav.addEventListener('click', () => {
      nav.classList.remove('is-open');
      menuButton.setAttribute('aria-expanded', 'false');
    });
  }

  setToday();
  hydrate();
})();
