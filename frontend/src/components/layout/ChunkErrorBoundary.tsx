import { Component, type ReactNode } from "react";

interface ChunkErrorBoundaryProps {
  resetKey: boolean;
  children: ReactNode;
}

interface ChunkErrorBoundaryState {
  crashed: boolean;
}

export default class ChunkErrorBoundary extends Component<
  ChunkErrorBoundaryProps,
  ChunkErrorBoundaryState
> {
  constructor(props: ChunkErrorBoundaryProps) {
    super(props);
    this.state = { crashed: false };
  }

  static getDerivedStateFromError(): ChunkErrorBoundaryState {
    return { crashed: true };
  }

  componentDidUpdate(prevProps: ChunkErrorBoundaryProps) {
    if (prevProps.resetKey !== this.props.resetKey && this.state.crashed) {
      this.setState({ crashed: false });
    }
  }

  render() {
    if (this.state.crashed) return null;
    return this.props.children;
  }
}
