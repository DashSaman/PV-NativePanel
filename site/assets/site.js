(() => {
  'use strict';

  const dataPath = '/data/articles.json';
  const fallbackImage = '/assets/news-fallback.svg';
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
      today.textContent = `${label} · مرور خبر، تصویر و ویدئو از منابع اصلی`;
    } catch (_) {
      today.textContent = 'مرور خبر، تصویر و ویدئو از منابع اصلی';
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

  const safeAssetUrl = (value) => {
    if (typeof value === 'string' && value.startsWith('/')) return value;
    const checked = safeUrl(value);
    return checked === '#' ? fallbackImage : checked;
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

  const makeImage = (article, width, height) => {
    const image = document.createElement('img');
    image.src = safeAssetUrl(article.image || fallbackImage);
    image.alt = article.image_alt || 'تصویر خبر';
    image.loading = 'lazy';
    image.width = width;
    image.height = height;
    image.addEventListener('error', () => { image.src = fallbackImage; }, { once: true });
    return image;
  };

  const meta = (article, source) => {
    const row = document.createElement('div');
    row.className = 'story-meta';
    const category = document.createElement('span');
    category.className = 'category';
    category.textContent = article.category || 'خبر';
    const date = document.createElement('span');
    date.textContent = article.published_label || 'منبع اصلی';
    row.append(category, date);
    if (source?.name) row.title = source.name;
    return row;
  };

  const makeFeedItem = (article, source) => {
    const card = document.createElement('article');
    card.className = 'feed-item';
    const image = makeImage(article, 340, 220);
    const body = document.createElement('div');
    body.className = 'feed-item__body';
    const title = document.createElement('h3');
    title.append(makeLink(article.title, article.source_url));
    const summary = document.createElement('p');
    summary.textContent = article.summary || '';
    const footer = document.createElement('div');
    footer.className = 'card-footer';
    const sourceName = document.createElement('span');
    sourceName.textContent = source?.domain || 'منبع اصلی';
    footer.append(sourceName, makeLink('اصل خبر', article.source_url));
    body.append(meta(article, source), title, summary, footer);
    card.append(image, body);
    return card;
  };

  const makeSideStory = (article, source) => {
    const card = document.createElement('article');
    card.className = 'side-story';
    const visual = makeLink('', article.source_url, 'side-story__image');
    visual.append(makeImage(article, 600, 400));
    const body = document.createElement('div');
    body.className = 'side-story__body';
    const title = document.createElement('h2');
    title.append(makeLink(article.title, article.source_url));
    const summary = document.createElement('p');
    summary.textContent = article.summary || '';
    const chip = makeLink(source?.domain || 'منبع اصلی', article.source_url, 'source-chip');
    body.append(meta(article, source), title, summary, chip);
    card.append(visual, body);
    return card;
  };

  const makeStackItem = (article, source) => {
    const item = document.createElement('article');
    const image = makeImage(article, 240, 150);
    const body = document.createElement('div');
    const title = document.createElement('h3');
    title.append(makeLink(article.title, article.source_url));
    const summary = document.createElement('p');
    summary.textContent = article.summary || '';
    body.append(meta(article, source), title, summary);
    item.append(image, body);
    return item;
  };

  const makeGalleryCard = (article, source) => {
    const card = document.createElement('article');
    card.className = 'gallery-card';
    const link = makeLink('', article.source_url);
    link.append(makeImage(article, 600, 400));
    const caption = document.createElement('span');
    const sourceName = document.createElement('small');
    sourceName.textContent = source?.domain || 'منبع اصلی';
    const title = document.createElement('strong');
    title.textContent = article.title;
    caption.append(sourceName, title);
    link.append(caption);
    card.append(link);
    return card;
  };

  const makeMediaCard = (article, source) => {
    const card = document.createElement('article');
    card.className = 'media-card';
    const visual = makeLink('', article.source_url, 'media-card__visual');
    visual.append(makeImage(article, 480, 270));
    const play = document.createElement('span');
    play.className = 'play-button';
    play.setAttribute('aria-hidden', 'true');
    play.textContent = '▶';
    const duration = document.createElement('small');
    duration.textContent = article.duration_label || 'ویدئو';
    visual.append(play, duration);
    const body = document.createElement('div');
    const sourceName = document.createElement('span');
    sourceName.textContent = source?.domain || 'منبع اصلی';
    const title = document.createElement('h3');
    title.append(makeLink(article.title, article.source_url));
    body.append(sourceName, title);
    card.append(visual, body);
    return card;
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
    const image = makeImage(article, 905, 1280);
    image.loading = 'eager';
    const badge = document.createElement('span');
    badge.className = 'visual-badge';
    badge.textContent = 'پرونده رهبری';
    visual.append(image, badge);

    const body = document.createElement('div');
    body.className = 'hero-story__body';
    const title = document.createElement('h1');
    title.append(makeLink(article.title, article.source_url));
    const summary = document.createElement('p');
    summary.textContent = article.summary || '';
    const actions = document.createElement('div');
    actions.className = 'hero-actions';
    actions.append(makeLink('مشاهده منبع اصلی', article.source_url, 'button-link'));
    if (source?.homepage) actions.append(makeLink(source.domain, source.homepage, 'plain-link'));
    body.append(meta(article, source), title, summary, actions);
    target.replaceChildren(visual, body);
  };

  const renderLeader = (articles, sources) => {
    const target = document.getElementById('leader-grid');
    if (!target || !articles.length) return;
    const fragment = document.createDocumentFragment();
    articles.forEach((article) => {
      const item = document.createElement('article');
      const label = document.createElement('span');
      label.textContent = article.published_label || article.category;
      const title = document.createElement('h3');
      title.append(makeLink(article.title, article.source_url));
      const summary = document.createElement('p');
      summary.textContent = article.summary || '';
      const source = sources.get(article.source_id);
      if (source?.name) item.title = source.name;
      item.append(label, title, summary);
      fragment.append(item);
    });
    target.replaceChildren(fragment);
  };

  const renderMediaFeature = (article, source) => {
    const target = document.getElementById('media-feature');
    if (!target || !article) return;
    const shell = document.createElement('div');
    shell.className = 'video-shell';
    const video = document.createElement('video');
    video.controls = true;
    video.preload = 'metadata';
    video.poster = safeAssetUrl(article.image || fallbackImage);
    if (article.media_url) {
      const media = document.createElement('source');
      media.src = safeUrl(article.media_url);
      media.type = 'video/mp4';
      video.append(media);
    }
    shell.append(video);
    const body = document.createElement('div');
    body.className = 'media-feature__body';
    const title = document.createElement('h3');
    title.append(makeLink(article.title, article.source_url));
    const summary = document.createElement('p');
    summary.textContent = article.summary || '';
    body.append(meta(article, source), title, summary, makeLink('صفحه رسمی ویدئو', article.source_url, 'source-chip'));
    target.replaceChildren(shell, body);
  };

  const renderMultimedia = (articles, sources) => {
    if (!articles.length) return;
    const directVideo = articles.find((article) => article.media_type === 'video') || articles[0];
    renderMediaFeature(directVideo, sources.get(directVideo.source_id));
    const rest = articles.filter((article) => article.id !== directVideo.id);
    replaceList('media-grid', rest, sources, makeMediaCard);
  };

  const renderBreaking = (article, source) => {
    if (!article) return;
    const link = document.getElementById('breaking-link');
    if (link) {
      link.textContent = article.title;
      link.href = safeUrl(article.source_url);
    }
    const sourceLabel = document.querySelector('.breaking__source');
    if (sourceLabel && source?.domain) sourceLabel.textContent = source.domain;
  };

  const renderPayload = (payload) => {
    if (!payload || !Array.isArray(payload.sources) || !Array.isArray(payload.articles)) throw new Error('invalid local article cache');
    const sources = sourceMap(payload);
    const bySection = (section) => payload.articles.filter((article) => article.section === section);
    const breaking = bySection('breaking')[0];
    const hero = bySection('hero')[0];
    renderBreaking(breaking, sources.get(breaking?.source_id));
    renderHero(hero, sources.get(hero?.source_id));
    replaceList('featured-grid', bySection('featured'), sources, makeSideStory);
    replaceList('latest-grid', bySection('latest'), sources, makeFeedItem);
    renderLeader(bySection('leader'), sources);
    renderMultimedia(bySection('multimedia'), sources);
    replaceList('gallery-grid', bySection('gallery'), sources, makeGalleryCard);
    replaceList('politics-grid', bySection('politics'), sources, makeStackItem);
    replaceList('economy-grid', bySection('economy'), sources, makeStackItem);
    replaceList('international-grid', bySection('international'), sources, makeStackItem);
    if (updated && payload.updated_label) updated.textContent = payload.updated_label;
    if (status) status.textContent = 'منابع محلی به‌روز';
    document.documentElement.dataset.newsState = 'ready';
  };

  const renderFallback = () => {
    document.documentElement.dataset.newsState = 'fallback';
    if (status) status.textContent = 'نمایش نسخه پایدار ذخیره‌شده';
    if (updated) updated.textContent = 'نسخه پایدار';
  };

  const hydrate = async () => {
    try {
      const response = await fetch(dataPath, { cache: 'no-store' });
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
