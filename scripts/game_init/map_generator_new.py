import random
import math
import numpy as np
from opensimplex import OpenSimplex
import matplotlib.pyplot as plt
import matplotlib.colors as pltc
from scipy.stats import qmc
from scipy.spatial import Voronoi, voronoi_plot_2d
from matplotlib.patches import RegularPolygon


def offset_to_cube(col, row):
    q = col - (row - (row & 1)) // 2
    r = row
    s = -q - r
    return q, r, s


def cube_distance(cube1, cube2):
    return max(
        abs(cube1[0] - cube2[0]), abs(cube1[1] - cube2[1]), abs(cube1[2] - cube2[2])
    )


def hex_distance(p1, p2):
    return cube_distance(offset_to_cube(*p1), offset_to_cube(*p2))


def get_hex_neighbors(x, y):
    if y % 2 == 0:
        return [
            (x - 1, y),
            (x + 1, y),
            (x, y - 1),
            (x + 1, y - 1),
            (x, y + 1),
            (x + 1, y + 1),
        ]
    else:
        return [
            (x - 1, y),
            (x + 1, y),
            (x - 1, y - 1),
            (x, y - 1),
            (x - 1, y + 1),
            (x, y + 1),
        ]


class Island:
    def __init__(self, id):
        pass


class WorldGenerator:
    # world gen
    world_w = 128
    world_h = 128
    border = 8

    # island gen
    i_min = 12
    i_size = 4
    m_min = 3
    m_max = 8
    p_min = 9
    p_max = 14
    c_min = 15
    c_max = 20

    def __init__(self):
        self.world = np.ones((self.world_w, self.world_h), dtype=int)
        self.islands = []
        self.centers = []

    def generate_map(self):
        self.world = np.ones((self.world_w, self.world_h), dtype=int)
        self.islands = self.generate_islands()
        self.generate_sea()

    def generate_island_centers(self, k=30) -> list[tuple]:
        """
        Fast Poisson Disk Sampling in Arbitrary Dimensions
        """
        rng = np.random.default_rng()
        phys_w = self.world_w * math.sqrt(3)
        phys_h = self.world_h * 1.5
        max_dim = max(
            phys_w - 2 * self.border * math.sqrt(3), phys_h - 2 * self.border * 1.5
        )

        # radius in physical distance (i_min hex distance roughly means i_min * sqrt(3) distance)
        engine = qmc.PoissonDisk(
            d=2, radius=(self.i_min * 1.5) / max_dim, rng=rng, nccandidates=k
        )
        sample = engine.fill_space()

        centers = []
        for sx, sy in sample:
            # Map back to physical
            cx_phys = sx * (
                phys_w - 2 * self.border * math.sqrt(3)
            ) + self.border * math.sqrt(3)
            cy_phys = sy * (phys_h - 2 * self.border * 1.5) + self.border * 1.5

            # Map physical back to hex coordinates
            y = int(cy_phys / 1.5)
            x = int((cx_phys - (y % 2) * (math.sqrt(3) / 2)) / math.sqrt(3))

            cx = min(self.world_w - 1, max(0, x))
            cy = min(self.world_h - 1, max(0, y))
            centers.append((cx, cy))

        return centers

    def generate_island_shape(self, center):
        cx, cy = center
        radius = self.i_size

        valid_island = False
        c_tiles = []
        p_tiles = []
        m_tiles = []

        while not valid_island:
            land = set()
            noise = OpenSimplex(seed=random.randint(0, 999999))

            max_r = int(radius * 2)
            for x in range(max(0, cx - max_r), min(self.world_w, cx + max_r + 1)):
                for y in range(max(0, cy - max_r), min(self.world_h, cy + max_r + 1)):
                    d = hex_distance((cx, cy), (x, y))

                    px = x * math.sqrt(3) + (y % 2) * (math.sqrt(3) / 2)
                    py = y * 1.5

                    n1 = noise.noise2(px * 0.1, py * 0.1)
                    threshold = radius * (1 + 0.4 * n1)

                    if d < threshold:
                        land.add((x, y))

            # Differentiate B (boundary) and I (internal)
            boundary = []
            internal = []

            for x, y in land:
                is_boundary = False
                for nx, ny in get_hex_neighbors(x, y):
                    if (
                        nx < 0
                        or ny < 0
                        or nx >= self.world_w
                        or ny >= self.world_h
                        or (nx, ny) not in land
                    ):
                        is_boundary = True
                        break
                if is_boundary:
                    boundary.append((x, y))
                else:
                    internal.append((x, y))

            c_count = random.randint(self.c_min, self.c_max)
            p_count = random.randint(self.p_min, self.p_max)

            if len(boundary) >= c_count and len(internal) >= p_count:
                m_count = (len(boundary) - c_count) + (len(internal) - p_count)
                if self.m_min <= m_count <= self.m_max:
                    valid_island = True

                    random.shuffle(boundary)
                    c_tiles = boundary[:c_count]
                    m_tiles = boundary[c_count:]

                    random.shuffle(internal)
                    p_tiles = internal[:p_count]
                    m_tiles.extend(internal[p_count:])

        return c_tiles, p_tiles, m_tiles

    def generate_islands(self):
        islands = []
        self.centers = self.generate_island_centers()
        for center in self.centers:
            c_tiles, p_tiles, m_tiles = self.generate_island_shape(center)
            islands.append((c_tiles, p_tiles, m_tiles))

            for x, y in c_tiles:
                self.world[y][x] = 3
            for x, y in p_tiles:
                self.world[y][x] = 4
            for x, y in m_tiles:
                self.world[y][x] = 5

        return islands

    def generate_sea(self):
        land_types = {3, 4, 5}
        sea_tiles = []
        for y in range(self.world_h):
            for x in range(self.world_w):
                if self.world[y][x] == 1:
                    adj_land = False
                    for nx, ny in get_hex_neighbors(x, y):
                        if 0 <= nx < self.world_w and 0 <= ny < self.world_h:
                            if self.world[ny][nx] in land_types:
                                adj_land = True
                                break
                    if adj_land:
                        sea_tiles.append((x, y))

        for x, y in sea_tiles:
            self.world[y][x] = 2

    def show_map(self):
        fig, ax = plt.subplots(figsize=(10, 10))
        ax.set_aspect("equal")

        colors = {
            1: "royalblue",  # Ocean
            2: "cornflowerblue",  # Sea
            3: "navajowhite",  # Coast
            4: "forestgreen",  # Plains
            5: "dimgrey",  # Mountain
        }

        # Instead of iterating 128x128 we only plot patches where necessary, or all
        for y in range(self.world_h):
            for x in range(self.world_w):
                val = self.world[y][x]
                cx = x * math.sqrt(3) + (y % 2) * (math.sqrt(3) / 2)
                cy = y * 1.5

                # Plot hexagons
                hexcoords = RegularPolygon(
                    (cx, cy),
                    numVertices=6,
                    radius=1.0,
                    orientation=np.pi / 6,
                    facecolor=colors[val],
                    edgecolor="none",
                )
                ax.add_patch(hexcoords)

        for xi, yi in self.centers:
            cx = xi * math.sqrt(3) + (yi % 2) * (math.sqrt(3) / 2)
            cy = yi * 1.5
            plt.plot(cx, cy, "ko", markersize=2)

        ax.set_xlim(-1, self.world_w * math.sqrt(3) + 1)
        ax.set_ylim(-1, self.world_h * 1.5 + 1)
        plt.title("Hexbin World Map")
        plt.axis("off")
        plt.show()

    def run(self):
        self.generate_map()
        self.show_map()


def run():
    generator = WorldGenerator()
    generator.run()


if __name__ == "__main__":
    run()
