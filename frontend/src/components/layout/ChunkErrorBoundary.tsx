import { Component, type ReactNode } from "react";

interface ChunkErrorBoundaryProps {
  resetKey: boolean;
  children: ReactNode;
}

interface ChunkErrorBoundaryState {
  crashed: boolean;
  resetKey: boolean;
}

export default class ChunkErrorBoundary extends Component<
  ChunkErrorBoundaryProps,
  ChunkErrorBoundaryState
> {
  constructor(props: ChunkErrorBoundaryProps) {
    super(props);
    this.state = { crashed: false, resetKey: props.resetKey };
  }

  static getDerivedStateFromError(): Partial<ChunkErrorBoundaryState> {
    return { crashed: true };
  }

  static getDerivedStateFromProps(
    props: ChunkErrorBoundaryProps,
    state: ChunkErrorBoundaryState,
  ): Partial<ChunkErrorBoundaryState> | null {
    if (props.resetKey !== state.resetKey) {
      return { crashed: false, resetKey: props.resetKey };
    }
    return null;
  }

  render() {
    if (this.state.crashed) return null;
    return this.props.children;
  }
}
