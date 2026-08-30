import { Component, ErrorInfo, ReactNode } from "react";

type Props = { children: ReactNode };
type State = { failed: boolean };

export class ErrorBoundary extends Component<Props, State> {
  state: State = { failed: false };

  static getDerivedStateFromError(): State {
    return { failed: true };
  }

  componentDidCatch(_error: Error, _info: ErrorInfo) {
    // Raw exceptions are intentionally not rendered because they can contain
    // request or implementation details. Operational logs carry request IDs.
  }

  render() {
    if (!this.state.failed) return this.props.children;
    return <main className="fatal-boundary" dir="rtl"><section className="fatal-card" role="alert"><p className="eyebrow">Safe fallback</p><h1>نمایش پنل با خطا روبه‌رو شد</h1><p>اطلاعات حساس یا متن خام خطا نمایش داده نمی‌شود. صفحه را دوباره بارگذاری کنید؛ اگر مشکل ادامه داشت، خروجی <code>pvnaive doctor</code> برای بررسی عملیاتی مناسب است.</p><button onClick={() => window.location.reload()}>بارگذاری مجدد</button></section></main>;
  }
}
