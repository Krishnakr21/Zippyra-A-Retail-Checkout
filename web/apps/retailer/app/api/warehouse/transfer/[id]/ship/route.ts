import { NextResponse } from 'next/server';

export async function PUT(req: Request, { params }: { params: { id: string } }) {
  return NextResponse.json({
    transfer_id: params.id,
    status: 'IN_TRANSIT',
    message: 'Transfer shipped successfully',
  });
}
