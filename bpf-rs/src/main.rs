use std::mem::MaybeUninit;
use libbpf_rs::skel::OpenSkel;
use libbpf_rs::skel::Skel;
use libbpf_rs::skel::SkelBuilder;
use anyhow::Result;
use std::thread;

mod xdp_pass_bpf {
    include!(concat!(
        env!("CARGO_MANIFEST_DIR"),
        "/src/bpf/xdp_pass.skel.rs"
    ));
}

use xdp_pass_bpf::*;

fn main() -> Result<()> {
    // bpf_object__open_skeleton
    let mut skel_builder = XdpPassSkelBuilder::default();
    let mut open_object = MaybeUninit::uninit();
    let open_skel = skel_builder.open(&mut open_object).unwrap();

    // bpf_object__load_skeleton
    let mut skel = open_skel.load().unwrap();

    // attach
    let mut link = skel.progs.xdp_pass.attach_xdp(1).unwrap();
    //skel.links = XdpPassLinks {
    //    xdp_pass: Some(link),
    //};
    link.detach().unwrap();

    thread::park();

    Ok(())
}
