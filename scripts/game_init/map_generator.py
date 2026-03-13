import random, math
import numpy as np
from opensimplex import OpenSimplex
import matplotlib.pyplot as plt
import matplotlib.colors as pltc
from scipy.stats import qmc
from scipy.spatial import Voronoi, voronoi_plot_2d


def norm1_distance(a, b):
    return abs(a[0] - b[0]) + abs(a[1] - b[1])


def norm2_distance(a, b):
    return math.sqrt((a[0] - b[0]) ** 2 + (a[1] - b[1]) ** 2)


class WorldGenerator:
    world_w = 100
    world_h = 100
    border = 5
    i_min = 15
    i_size = 7
    m_min = 3
    m_max = 6
    p_min = 9
    p_max = 12
    c_min = 15
    c_max = 18

    def __init__(self):
        self.world = [[0 for _ in range(self.world_w)] for _ in range(self.world_h)]

    def generate_map(self):
        self.centers = self.generate_island_centers(self.border)
        for center in self.centers:
            x, y = center
            self.world[y][x] = 5

            land = self.generate_island_shape(center)
            for l in land:
                x, y = l
                self.world[y][x] = 5

    def generate_island_centers(self, margin, k=30) -> list[tuple]:
        """
        Fast Poisson Disk Sampling in Arbitrary Dimensions

        Args:
            k (int, optional): Samples before rejection. Defaults to 30.
        Returns:
            list[tuple]: Valid center list.
        """

        rng = np.random.default_rng()
        engine = qmc.PoissonDisk(
            d=2, radius=self.i_min / max(self.world_w-2*margin, self.world_h-2*margin), 
            rng=rng, ncandidates=k
        )
        sample = engine.fill_space()

        centers = []
        for x, y in sample:
            cx = int(x * self.world_w) + self.border
            cy = int(y * self.world_h) + self.border

            cx = self.world_w if cx > self.world_w else cx
            cy = self.world_h if cy > self.world_h else cy
            centers.append((cx, cy))

        return centers

    def generate_island_shape(self, center):
        cx, cy = center
        radius = 4 #random.randint(3,5)
        valid_island = False

        while not valid_island:
            land = set()
            noise = OpenSimplex(seed=random.randint(0, 999999))
            for x in range(max(0, cx-radius*2), min(self.world_w, cx+radius*2)):
                for y in range(max(0, cy-radius*2), min(self.world_h, cy+radius*2)):

                    d = math.sqrt((x-cx)**2 + (y-cy)**2)

                    n1 = noise.noise2(x*0.01, y*0.05)
                    n2 = noise.noise2(x*0.1, y*0.15)
                    # n3 = noise.noise2(x*0.25, y*0.25)
                    nval = n1 + n2

                    threshold = radius * (1 + 0.4*nval)

                    if d < threshold:
                        land.add((x, y))
            min_size = self.m_min + self.p_min + self.c_min
            max_size = self.m_max + self.p_max + self.c_max
            if min_size <= len(land) <= max_size:
                valid_island = True
            
        return land

    def generate_islands(self):
        pass

    def show_map(self):
        cvals = [1, 2, 3, 4, 5]
        colors = [
            "royalblue",
            "cornflowerblue",
            "navajowhite",
            "forestgreen",
            "black"# "dimgrey",
        ]

        norm = plt.Normalize(min(cvals), max(cvals))
        tuples = list(zip(map(norm, cvals), colors))
        cmap = pltc.LinearSegmentedColormap.from_list("", tuples)
        cmap = pltc.ListedColormap(colors)

        fig, ax = plt.subplots()

        # plot island centers with minimum distance
        for xi, yi in self.centers:
            circle = plt.Circle((xi, yi), radius=self.i_min/2-1, fill=False)    
            ax.add_artist(circle)

        # plot world map
        ax.imshow(self.world, cmap=cmap, vmin=0.5, vmax=5.5)
        ax.set_xlim(0, self.world_w-1)
        ax.set_ylim(0, self.world_h-1)

        # plot voronoi regions
        vor = Voronoi(self.centers)
        voronoi_plot_2d(vor, ax=ax, show_vertices=False, line_colors='black')

        # plt.colorbar()
        plt.title("World Map")
        plt.show()

    def run(self):
        self.generate_map()
        self.generate_islands()
        self.show_map()


def run():
    generator = WorldGenerator()
    generator.run()


if __name__ == "__main__":
    run()
