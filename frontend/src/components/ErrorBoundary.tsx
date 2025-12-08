import React, { Component, ErrorInfo, ReactNode } from 'react'

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
  }

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo)
  }

  public render() {
    if (this.state.hasError) {
      return (
        this.props.fallback || (
          <div className="p-4 m-4 bg-red-500/10 border border-red-500 rounded-lg">
            <h2 className="text-red-500 font-bold mb-2">Ошибка загрузки страницы</h2>
            <p className="text-sm text-red-400 font-mono">
              {this.state.error?.message || 'Неизвестная ошибка'}
            </p>
            <pre className="mt-2 text-xs text-red-300 overflow-auto max-h-40">
              {this.state.error?.stack}
            </pre>
            <button
              onClick={() => this.setState({ hasError: false, error: null })}
              className="mt-4 px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600"
            >
              Попробовать снова
            </button>
          </div>
        )
      )
    }

    return this.props.children
  }
}
