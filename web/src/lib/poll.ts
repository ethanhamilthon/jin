// every runs fn now and then once a second while the page is visible, until the
// returned function is called. A run that is still going is not started twice.
export function every(fn: () => Promise<unknown> | unknown, ms = 1000): () => void {
  let busy = false;
  const run = async () => {
    if (busy || document.hidden) return;
    busy = true;
    try {
      await fn();
    } finally {
      busy = false;
    }
  };
  const timer = setInterval(run, ms);
  return () => clearInterval(timer);
}
