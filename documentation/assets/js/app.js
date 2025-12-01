
document.addEventListener("DOMContentLoaded", () => {
  const routes = {
    docs: "pages/docs.html",
    equipo: "pages/equipo.html",
    apache: "pages/apache.html",
    ftp: "pages/ftp.html",
    ssh: "pages/ssh.html",
    securemonitor: "pages/securemonitor.html",
  };

  const appMain = document.getElementById("app-main");
  const navLinks = Array.from(document.querySelectorAll(".nav-link[data-page]"));

  function setActiveNav(pageKey) {
    navLinks.forEach(link => {
      const isActive = link.dataset.page === pageKey;
      link.classList.toggle("nav-link--primary", isActive);
    });
  }

  function updateTitle(pageKey) {
    const link = navLinks.find(l => l.dataset.page === pageKey);
    const titleFromLink = link?.dataset.title;
    if (titleFromLink) {
      document.title = titleFromLink;
    } else {
      document.title = "SecureMonitor · Documentación";
    }
  }

  async function loadPage(pageKey, { pushState = true } = {}) {
    const route = routes[pageKey] || routes.docs;

    try {
      const res = await fetch(route, { cache: "no-cache" });
      if (!res.ok) {
        throw new Error("HTTP " + res.status);
      }

      const html = await res.text();
      const tmp = document.createElement("div");
      tmp.innerHTML = html;

      const newMain = tmp.querySelector("main.layout");
      if (!newMain) {
        throw new Error('No se encontró <main class="layout"> en ' + route);
      }

      newMain.classList.add("page-fade-in");

      appMain.innerHTML = "";
      appMain.appendChild(newMain);

      newMain.addEventListener(
        "animationend",
        () => {
          newMain.classList.remove("page-fade-in");
        },
        { once: true }
      );

      setActiveNav(pageKey);
      updateTitle(pageKey);

      if (pushState) {
        history.pushState({ page: pageKey }, "", "#" + pageKey);
      }
    } catch (err) {
      console.error(err);
      appMain.innerHTML = `
        <main class="layout">
          <section class="content">
            <div style="padding: 2rem; color: #fecaca; font-size: 0.9rem;">
              Ocurrió un error al cargar la página <strong>${pageKey}</strong>.<br/>
              Detalle: ${err.message}
            </div>
          </section>
        </main>
      `;
    }
  }

  navLinks.forEach(link => {
    link.addEventListener("click", (e) => {
      e.preventDefault();
      const pageKey = link.dataset.page;
      if (!pageKey) return;

      if (link.classList.contains("nav-link--primary")) return;

      loadPage(pageKey, { pushState: true });
    });
  });

  document.addEventListener("click", (e) => {
    const sidebarLink = e.target.closest(".sidebar-link");
    if (!sidebarLink) return;

    const href = sidebarLink.getAttribute("href");
    if (!href || !href.startsWith("#")) return;

    e.preventDefault();
    e.stopPropagation();

    const target = document.querySelector(href);
    if (target) {
      target.scrollIntoView({ behavior: "smooth", block: "start" });

      const currentState = history.state || {};
      history.replaceState(currentState, "", href);
    }
  });

  window.addEventListener("popstate", (event) => {
    const pageKey = event.state?.page;
    if (!pageKey || !routes[pageKey]) {
      return;
    }
    loadPage(pageKey, { pushState: false });
  });

  function getInitialPage() {
    const hash = window.location.hash.replace("#", "").trim();
    if (routes[hash]) return hash;
    return "docs";
  }

  const initialPage = getInitialPage();
  loadPage(initialPage, { pushState: false });
});
