import type { NextPage } from "next";
import { Col, Container, Figure, Row } from "react-bootstrap";

import Avatar from "react-avatar";

import Image from 'next/image'

import SellComputeImage from "../../public/we-will-sell-compute.png";
import ShareComputeImage from "../../public/we-will-share-compute.png";

const About: NextPage = () => {
  return (
    <Container>
      <h1>DAOnetes</h1>
      <div className="wallOfText">
        Nathan LeClaire and Sven Dowideit founded Crypto Workbench in 2022
        to build an ad-hoc distributed computing solution - DAOnetes.
        We met working at Docker in 2014 and bonded over our passion for
        unlocking computing’s “bicycle for the mind”. We’re both systems nerds
        with a blend of technical and customer facing experience.
        <Row>
          <Col></Col>
          <Col></Col>
          <Col>
            <Figure className="glass">
              <Avatar
                name="Nathan LeClaire"
                githubHandle="nathanleclaire"
                twitterHandle="dotpem"
                size="150"
              />
              <Figure.Caption>
                Nathan LeClaire:
                <br /> Docker, Honeycomb, Author for O’Reilly.
              </Figure.Caption>
            </Figure>
          </Col>
          <Col></Col>

          <Col>
            <Figure className="glass">
              <Avatar
                name="Sven Dowideit"
                githubHandle="SvenDowideit"
                twitterHandle="SvenDowideit"
                size="150"
              />
              <Figure.Caption>
                Sven Dowideit:
                <br /> Docker, Rancher, Portainer, CSIRO.
              </Figure.Caption>
            </Figure>
          </Col>
          <Col></Col>
          <Col></Col>
        </Row>
        Docker’s dream was to make a programmable internet - a global
        supercomputer. Unfortunately, we're not there yet, and for now,
        the industry landed on centralized solutions like Kubernetes.
      </div>
      <div className="wallOfText">
        <h2>Building Decentralized Supercomputers</h2>
        As we explored modern blockchain systems, we realized that it offers
        fundamental pieces that were missing in 2015. For instance,
        bootstrapping a cluster always ended up leaning on some centralized data
        store like AWS’s DynamoDB. Another that we ran into was identity — if
        users were going to share computers, they needed some way to recognize
        and delegate authority. Last but not least, connecting otherwise
        disparate environments and sharing specifications for running your apps
        was impossible without a centralized service.
      </div>
      <div className="wallOfText">
        We think web3 can roll the ball forward on these issues. Things like
        identity, auditability, and payments are built right in. Likewise, it’s
        a shared data store in the sky — so computers can discover each other
        and share information without having to trust some SaaS service.
        Suddenly, better collaboration (like accessing a container running on
        your teammate’s computer) and truly decentralized workloads seem
        possible, since there’s an answer to Kubernetes’ reliance on a
        centralized data store and runtime environment. So, for the first time,
        we can offer an experience where you execute a command or use an API
        without worrying about the details of precisely how and where it’s
        operated.
      </div>
      <div className="wallOfText">
        <h3>But does anyone want this? And why now? And why us?</h3>
        Why would anyone use a model like this when they can just use AWS? We
        think this is well worth consideration. Is it because it could offer
        cheaper compute? No, we don’t believe that’s a good strategy — cost is
        not the main driver of that decision. Is it because it’s more private or
        more secure? Not really. If you’re renting compute from a random seller,
        how can you know that you can trust them or intermediaries with your
        secrets like API keys, data flow, etc.? These are issues we have with
        the projects we’ve seen attempt to tackle decentralized compute networks
        in crypto today.
        <Row>
          <Col></Col>
          <Col></Col>
          <Col>
            <Figure className="glass">
              <Image
                alt="compute for sale"
                src={SellComputeImage.src}
                height={SellComputeImage.height}
                width={SellComputeImage.width} />
              <Figure.Caption>
                Is it decentralized? Or just built on decentralized technology?
              </Figure.Caption>
            </Figure>
          </Col>
          <Col></Col>
          <Col></Col>
        </Row>
        The big cloud providers aren’t going anywhere. So why bother? We think
        this is still a massive opportunity because we see that the world is
        becoming a place where each of us is swimming in a sea of trillions of
        localized computing devices. Workloads will need to be running *where
        the things are happening*. We should all be able to plug in to the mesh
        and utilize it no matter our location. You won’t always be calling out
        to centralized clouds and lugging your powerful Macbook around in the
        future, you’ll be tapping in to the network of compute all around you.
      </div>
      <div className="wallOfText">
        So, we will build a system for doing this. We jokingly call this system
        “DAOnetes”. (DAO+Kubernetes) It will enable you to plug in and out of
        arbitrary clusters by leveraging the consistency that containers offer,
        alongside the ability of the blockchain to trustlessly transact, port
        identity, and form groups for delegation of responsibility.
      </div>
      <div className="wallOfText">
        You will be able to bring your own cloud. Eventually, operations experts
        will bring theirs, and sell services for you to use, just like AWS does,
        but operated transparently.
        <Row>
          <Col></Col>
          <Col></Col>
          <Col>
            <Figure className="glass">
              <Image
                alt="compute for sharing"
                src={ShareComputeImage.src}
                height={ShareComputeImage.height}
                width={ShareComputeImage.width} />
              <Figure.Caption>
                The vision of Crypto Workbench is to create truly decentralized
                clusters that people can plug in and out of at will, including
                with small scale nodes and edge devices.
              </Figure.Caption>
            </Figure>
          </Col>
          <Col></Col>
          <Col></Col>
        </Row>
        Why us? Because we are experts in developer experience and making
        technology accessible, which will be critical to bootstrap the network.
        And we have a fire burning in our bellies to bring this vision to the
        world. This is the system that we want, and we saw no other way to build
        it than to start this company.
      </div>
    </Container>
  )
};

export default About;
